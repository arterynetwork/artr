package app

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/delegating"
	dTypes "github.com/arterynetwork/artr/x/delegating/types"
	"github.com/arterynetwork/artr/x/noding"
	nodingTypes "github.com/arterynetwork/artr/x/noding/types"
	"github.com/arterynetwork/artr/x/profile"
	"github.com/arterynetwork/artr/x/referral"
	refTypes "github.com/arterynetwork/artr/x/referral/types"
	schTypes "github.com/arterynetwork/artr/x/schedule/types"
	"github.com/arterynetwork/artr/x/storage"
)

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func Chain(handlers ...upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}

func CliWarningUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {
	fmt.Println(`
╔═════════════════════════════════════════════════════════════╗
║ PLEASE MAKE YOU SURE YOU HAVE UPGRADED CLI CLIENT AS WELL ! ║
╚═════════════════════════════════════════════════════════════╝`,
	)
}

func RefreshStatus(k referral.Keeper, status referral.Status) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		k.Iterate(ctx, func(_ sdk.AccAddress, r *referral.DataRecord) (changed, checkForStatusUpdate bool) {
			return false, r.Status == status
		})
	}
}

func RestoreTrafficLimit(k storage.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		k.Iterate(ctx, []byte{0x01} /* limitPrefix */, func(key sdk.AccAddress, value *uint64) (changed bool) {
			if *value < 5*util.GBSize {
				*value = 5 * util.GBSize
				return true
			} else {
				return false
			}
		})
	}
}

func ScheduleCompression(k referral.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		k.Iterate(ctx, func(acc sdk.AccAddress, r *referral.DataRecord) (changed, checkForStatusUpdate bool) {
			var compressionAt int64
			if !r.Active && r.CompressionAt < ctx.BlockHeight() {
				if r.CompressionAt == -1 {
					createdAt, ok := util.CreatedAt[acc.String()]
					if ok {
						compressionAt = createdAt + referral.CompressionPeriod*(1+(ctx.BlockHeight()-createdAt)/referral.CompressionPeriod)
					} else {
						compressionAt = ctx.BlockHeight() + referral.CompressionPeriod
					}
				} else {
					compressionAt = r.CompressionAt + referral.CompressionPeriod*(1+(ctx.BlockHeight()-r.CompressionAt)/referral.CompressionPeriod)
				}
				r.CompressionAt = compressionAt
				changed = true
				if err := k.ScheduleCompression(ctx, acc, compressionAt); err != nil {
					panic(err)
				}
			}
			return changed, false
		})
	}
}

func CountRevoking(ak auth.AccountKeeper, rk referral.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		ak.IterateAccounts(ctx, func(account authTypes.Account) (stop bool) {
			if account.GetCoins().AmountOf(util.ConfigRevokingDenom).IsPositive() {
				if err := rk.OnBalanceChanged(ctx, account.GetAddress()); err != nil {
					panic(err)
				}
			}
			return false
		})
	}
}

func InitializeTransitionCost(k referral.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting InitializeTransitionCost...")
		var pz referral.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, refTypes.KeyTransitionCost) {
				pz.TransitionCost = refTypes.DefaultTransitionCost
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		logger.Debug("Finished InitializeTransitionCost", "params", pz)
		k.SetParams(ctx, pz)
	}
}

func ClearInvalidNicknames(ak auth.AccountKeeper, pk profile.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		originals := make(map[string]sdk.AccAddress)

		ak.IterateAccounts(ctx, func(acc authTypes.Account) (stop bool) {
			address := acc.GetAddress()
			p := pk.GetProfile(ctx, address)
			if p == nil {
				return
			}
			nickname := p.Nickname

			if _, ok := originals[nickname]; !ok {
				if err := pk.ValidateProfileNickname(ctx, address, nickname); err == nil {
					return
				} else if err == profile.ErrNicknameAlreadyInUse {
					originals[nickname] = pk.GetProfileAccountByNickname(ctx, nickname)
				}
			}

			p.Nickname = ""
			if err := pk.SetProfile(ctx, address, *p); err != nil {
				// This cannot happen because the nickname is empty.
				panic(err)
			}

			return
		})

		for nick, address := range originals {
			p := pk.GetProfile(ctx, address)
			p.Nickname = nick
			if err := pk.SetProfile(ctx, address, *p); err != nil {
				// This cannot happen because we've just cleaned all up.
				panic(err)
			}
		}
	}
}

func InitializeMinDelegate(k delegating.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting InitializeMinDelegate...")
		var pz delegating.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, dTypes.KeyMinDelegate) {
				pz.MinDelegate = dTypes.DefaultMinDelegate
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		logger.Debug("Finished InitializeMinDelegate", "params", pz)
		k.SetParams(ctx, pz)
	}
}

func RebuildTeamCoinsCache(rk referral.Keeper, ak auth.AccountKeeper) upgrade.UpgradeHandler {
	type teamCoinsCacheData struct {
		total     []sdk.Int
		delegated []sdk.Int
		parent    string
	}

	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting RebuildTeamCoinsCache...")
		data := make(map[string]teamCoinsCacheData, 100_000)

		logger.Debug("    gathering data...")
		rk.Iterate(ctx, func(acc sdk.AccAddress, r *referral.DataRecord) (changed, checkForStatusUpdate bool) {
			key := acc.String()
			coins := ak.GetAccount(ctx, acc).GetCoins()
			var parent string
			if r.Referrer != nil {
				parent = r.Referrer.String()
			}

			data[key] = teamCoinsCacheData{
				total: []sdk.Int{
					coins.AmountOf(util.ConfigMainDenom).Add(coins.AmountOf(util.ConfigDelegatedDenom)).Add(coins.AmountOf(util.ConfigRevokingDenom)),
					sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
					sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
				},
				delegated: []sdk.Int{
					coins.AmountOf(util.ConfigDelegatedDenom),
					sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
					sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
				},
				parent: parent,
			}
			return false, false
		})
		logger.Debug("    calculating...", "count", len(data))
		for key, x := range data {
			c := x.total[0]
			d := x.delegated[0]

			var anc = key
			for i := 1; i <= 10; i++ {
				anc = data[anc].parent
				if anc == "" {
					break
				}

				y := data[anc]
				y.total[i] = y.total[i].Add(c)
				y.delegated[i] = y.delegated[i].Add(d)
			}
		}
		logger.Debug("    applying...")
		rk.Iterate(ctx, func(acc sdk.AccAddress, r *referral.DataRecord) (changed, checkForStatusUpdate bool) {
			key := acc.String()
			x := data[key]

			for i := 0; i <= 10; i++ {
				if !r.Coins[i].Equal(x.total[i]) {
					changed = true
					checkForStatusUpdate = true
					break
				}
			}
			if !changed {
				for i := 0; i <= 10; i++ {
					if !r.Delegated[i].Equal(x.delegated[i]) {
						changed = true
						break
					}
				}
			}

			if changed {
				logger.Debug("    record fixed",
					"address", acc,
					"coins_0", r.Coins,
					"coins", x.total,
					"delegated_0", r.Delegated,
					"delegated", x.delegated,
				)
				copy(r.Coins[:], x.total)
				copy(r.Delegated[:], x.delegated)
			}
			return
		})
		logger.Debug("    all done")
	}
}

func InitializeNodingLottery(k noding.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting InitializeNodingLottery...")
		var pz noding.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, nodingTypes.KeyLotteryValidators) {
				pz.LotteryValidators = nodingTypes.DefaultLotteryValidators
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Debug("Finished InitializeNodingLottery", "params", pz)
	}
}

func CheckStatusIndex(k referral.Keeper, indexStoreKey sdk.StoreKey) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting CheckStatusIndex...")
		k.Iterate(ctx, func(acc sdk.AccAddress, r *refTypes.R) (changed, checkForStatusUpdate bool) {
			if r.Status < referral.StatusBusinessman {
				return false, false
			}

			store := ctx.KVStore(indexStoreKey)
			key := make([]byte, len([]byte(acc))+1)
			copy(key[1:], acc)

			for status := referral.StatusBusinessman; status < r.Status; status++ {
				key[0] = uint8(status)
				if store.Has(key) {
					logger.Info("Clear wrong entry", "status", status, "acc", acc.String())
					store.Delete(key)
				}
			}
			key[0] = uint8(r.Status)
			if !store.Has(key) {
				logger.Info("Add missing entry", "status", r.Status, "acc", acc.String())
				store.Set(key, []byte{0x01})
			}
			return false, false
		})
	}
}

func InitializeNodingMinStatus(k noding.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting InitializeNodingMinStatus...")
		var pz noding.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, nodingTypes.KeyMinStatus) {
				pz.MinStatus = nodingTypes.DefaultMinStatus
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Debug("Finished InitializeNodingMinStatus", "params", pz)
	}
}

func ShardCompression(rk referral.Keeper, cdc *codec.Codec, rKey, schKey sdk.StoreKey) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting ShardCompression...")
		const MAX_PER_BLOCK = 10

		var (
			store = cachekv.NewStore(ctx.KVStore(schKey)) // CacheKV allows us to mutate the store while iterating over it.

			carryOn   []schTypes.Task
			carryOnTo uint64
		)

		var updateReferral func(sdk.AccAddress, uint64)
		{
			store := ctx.KVStore(rKey)
			updateReferral = func(acc sdk.AccAddress, height uint64) {
				var refData referral.DataRecord
				if err := cdc.UnmarshalBinaryLengthPrefixed(store.Get(acc), &refData); err != nil {
					panic(err)
				}
				refData.CompressionAt = int64(height)
				bz, err := cdc.MarshalBinaryLengthPrefixed(refData)
				if err != nil {
					panic(err)
				}
				store.Set(acc, bz)
			}
		}

		var bunchUpdateReferral = func(list schTypes.Schedule, height uint64) {
			for _, item := range list {
				updateReferral(item.Data, height)
			}
		}

		var scheduleNew = func(list schTypes.Schedule, height uint64, stop uint64) (remain schTypes.Schedule) {
			key := make([]byte, 8)
			var now schTypes.Schedule
			for ; list != nil && height != stop; height += 1 {
				binary.BigEndian.PutUint64(key, height)
				if len(carryOn) <= MAX_PER_BLOCK {
					now = carryOn
					carryOn = nil
				} else {
					now = carryOn[:MAX_PER_BLOCK]
					carryOn = carryOn[MAX_PER_BLOCK:]
				}
				bz, err := cdc.MarshalBinaryBare(now)
				if err != nil {
					panic(err)
				}
				store.Set(key, bz)
				bunchUpdateReferral(now, height)
			}
			return list
		}

		it := store.Iterator(nil, nil)
		for ; it.Valid(); it.Next() {
			var (
				all,
				head,
				tail schTypes.Schedule
				trash,
				n int
				height = binary.BigEndian.Uint64(it.Key())
			)
			if err := cdc.UnmarshalBinaryBare(it.Value(), &all); err != nil {
				panic(err)
			}
			carryOn = scheduleNew(carryOn, carryOnTo, height)
			for _, item := range carryOn {
				if n++; n > MAX_PER_BLOCK {
					tail = append(tail, item)
				} else {
					head = append(head, item)
					updateReferral(item.Data, height)
				}
			}
			for _, item := range all {
				if item.HandlerName == referral.CompressionHookName {
					h, err := rk.GetCompressionBlockHeight(ctx, item.Data)
					if err != nil {
						panic(err)
					}
					if h != int64(height) {
						trash++
						continue
					}
					if n++; n > MAX_PER_BLOCK {
						tail = append(tail, item)
						continue
					}
				}
				head = append(head, item)
			}
			if trash == 0 && carryOn == nil && tail == nil {
				continue
			}

			if trash != 0 {
				logger.Info(
					"Dropping obsolete compressions",
					"height", height,
					"count", trash,
				)
			}
			if head == nil {
				store.Delete(it.Key())
			} else {
				bz, err := cdc.MarshalBinaryBare(head)
				if err != nil {
					panic(err)
				}
				store.Set(it.Key(), bz)
			}

			carryOn = tail
			if tail != nil {
				logger.Info(
					"Too many compressions, rescheduling some to the next block",
					"height", height,
					"count", n,
				)
				carryOnTo = height + 1
			}
		}
		it.Close()
		if scheduleNew(carryOn, carryOnTo, 0) != nil {
			panic("Unscheduled compressions remain")
		}
		store.Write()

		logger.Debug("Finished ShardCompression")
	}
}
