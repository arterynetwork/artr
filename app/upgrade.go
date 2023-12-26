package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	bankT "github.com/arterynetwork/artr/x/bank/types"
	delegatingK "github.com/arterynetwork/artr/x/delegating/keeper"
	delegatingT "github.com/arterynetwork/artr/x/delegating/types"
	"github.com/arterynetwork/artr/x/noding"
	referralK "github.com/arterynetwork/artr/x/referral/keeper"
	referralT "github.com/arterynetwork/artr/x/referral/types"
	scheduleK "github.com/arterynetwork/artr/x/schedule/keeper"
	scheduleT "github.com/arterynetwork/artr/x/schedule/types"
	votingKeeper "github.com/arterynetwork/artr/x/voting/keeper"
	votingTypes "github.com/arterynetwork/artr/x/voting/types"
)

func Chain(handlers ...upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func RecalculateActiveReferrals(k referralK.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RecalculateActiveReferrals ...")
		k.Iterate(ctx, func(acc string, info *referralT.Info) (changed, checkForStatusUpdate bool) {
			arl := make([]string, 0, len(info.Referrals))
			ars := make(map[string]bool, len(info.Referrals))
			for _, rAddr := range info.Referrals {
				rInfo, err := k.Get(ctx, rAddr)
				if err != nil {
					logger.Error("Account %s not found", rAddr)
					panic(err)
				}
				if rInfo.Active {
					arl = append(arl, rAddr)
					ars[rAddr] = true
				}
			}
			oars := make(map[string]bool, len(info.ActiveReferrals))
			for _, addr := range info.ActiveReferrals {
				if !ars[addr] {
					changed = true
					break
				}
				oars[addr] = true
			}
			if !changed {
				for _, addr := range arl {
					if !oars[addr] {
						changed = true
						break
					}
				}
			}
			if changed {
				logger.Debug("... %s: %v → %v", acc, info.ActiveReferrals, arl)
				info.ActiveReferrals = arl
				return true, true
			}
			return false, false
		})
		logger.Info("... done")
	}
}

func ScheduleBanishment(rk referralK.Keeper, bk bank.Keeper, rKey, sKey sdk.StoreKey, cdc codec.BinaryMarshaler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting ScheduleBanishment ...")

		{
			var (
				ms  = ctx.MultiStore().CacheMultiStore()
				ctx = ctx.WithMultiStore(ms)

				sStore = ctx.KVStore(sKey)
				sIt    = sStore.Iterator(nil, nil)

				rStore = ctx.KVStore(rKey)

				dd = bk.GetParams(ctx).DustDelegation
			)
			logger.Info("... fixing schedule ...")
			for ; sIt.Valid(); sIt.Next() {
				var sch scheduleT.Schedule
				cdc.MustUnmarshalBinaryBare(sIt.Value(), &sch)

				changed := false
				tasks := make([]scheduleT.Task, 0, len(sch.Tasks))

			TasksForAMomentInTime:
				for _, task := range sch.Tasks {
					if task.HandlerName == referralK.BanishHookName {
						var addr sdk.AccAddress

						switch len(task.Data) {
						case AddrLen:
							addr = task.Data

							before := string(task.Data)
							after := sdk.AccAddress(task.Data).String()

							task.Data = []byte(after)
							changed = true

							logger.Debug("... ... restoring task data",
								"before", before,
								"after", after,
								"t", task.Time.String(),
							)
						case 43: // Bech32
							var err error
							if addr, err = sdk.AccAddressFromBech32(string(task.Data)); err != nil {
								logger.Error("Length is OK, but still cannot parse",
									"acc", string(task.Data),
									"err", err,
								)
							}
						default:
							logger.Error("Unexpected length", "acc", string(task.Data), "len", len(task.Data))
						}

						var r referralT.Info
						cdc.MustUnmarshalBinaryBare(rStore.Get(task.Data), &r)
						var wrong = r.Active || r.Banished
						if r.CompressionAt != nil && r.CompressionAt.After(ctx.BlockTime()) {
							var found bool

							var sch scheduleT.Schedule
							key := make([]byte, 8)
							binary.BigEndian.PutUint64(key, uint64(r.CompressionAt.UnixNano()))
							if schBz := sStore.Get(key); schBz != nil {
								cdc.MustUnmarshalBinaryBare(schBz, &sch)
								for _, t := range sch.Tasks {
									if t.HandlerName == referralK.CompressionHookName && string(t.Data) == addr.String() {
										found = true
										break
									}
								}
							}

							if found {
								wrong = true
							} else {
								logger.Debug("... ... deleting   CompressionAt",
									"acc", addr.String(),
									"was", r.CompressionAt.String(),
								)
								r.CompressionAt = nil
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							}
						}
						wrong = wrong || len(r.Delegated) > 0 && r.Delegated[0].Int64() > dd

						if wrong {
							logger.Debug("... ... unscheduling",
								"acc", addr.String(),
								"t", task.Time.String(),
							)
							if r.BanishmentAt != nil {
								logger.Debug("... ... erasing    BanishmentAt ",
									"acc", addr.String(),
									"was", r.BanishmentAt.String(),
								)
								r.BanishmentAt = nil
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							}
							changed = true
							continue TasksForAMomentInTime
						} else {
							if r.BanishmentAt == nil {
								logger.Debug("... ... recovering BanishmentAt ",
									"acc", addr.String(),
									"val", task.Time.String(),
								)
								r.BanishmentAt = &task.Time
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							} else if !r.BanishmentAt.Equal(task.Time) {
								logger.Debug("... ... fixing     BanishmentAt ",
									"acc", addr.String(),
									"from", r.BanishmentAt.String(),
									"to", task.Time.String(),
								)
								r.BanishmentAt = &task.Time
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							}
						}
					}
					tasks = append(tasks, task)
				}

				if changed {
					if len(tasks) == 0 {
						sStore.Delete(sIt.Key())
					} else {
						sch.Tasks = tasks
						sStore.Set(sIt.Key(), cdc.MustMarshalBinaryBare(&sch))
					}
				}
			}
			sIt.Close()

			logger.Info("... fixing referral ...")
			var (
				rIt = rStore.Iterator(nil, nil)
			)
			for ; rIt.Valid(); rIt.Next() {
				acc := string(rIt.Key())
				if !strings.HasPrefix(acc, "artr1") {
					logger.Debug("... ... deleting from referral", "acc", acc)
					rStore.Delete(rIt.Key())
					continue
				}

				var r referralT.Info
				cdc.MustUnmarshalBinaryBare(rIt.Value(), &r)
				if r.Banished && r.Active {
					logger.Info("... ... banished yet active, restoring",
						"acc", acc,
					)
					if err := rk.ComeBack(ctx, acc); err != nil {
						logger.Error("Cannot restore unfairly banished account", "acc", acc, "err", err)
					}
				}
			}
			rIt.Close()

			logger.Debug("... persisting multistore")
			ms.Write()
		}
		logger.Info("... ScheduleBanishment done!")
	}
}

func InitPollPeriodParam(k votingKeeper.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitPollPeriodParam ...")

		var pz votingTypes.Params
		paramspace.Get(ctx, votingTypes.KeyParamVotingPeriod, &pz.VotingPeriod)
		pz.PollPeriod = pz.VotingPeriod

		k.SetParams(ctx, pz)
		logger.Info("... InitPollPeriodParam done!", "params", pz)
	}
}

func ForceOnStatusChangedCallback(k noding.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting ForceOnStatusChangedCallback ...")

		if validators, err := k.GetActiveValidatorList(ctx); err != nil {
			logger.Error("Cannot get active validator list", "err", err)
		} else {
			for _, acc := range validators {
				if ok, _, _, err := k.IsQualified(ctx, acc); err != nil {
					logger.Error("Cannot check qualification", "err", err, "acc", acc)
				} else if !ok {
					logger.Info("... switching off validator", "acc", acc)
					if err = k.SwitchOff(ctx, acc); err != nil {
						logger.Error("Cannot switch off validator", "err", err, "acc", acc)
					}
				}
			}
		}

		logger.Info("... ForceOnStatusChangedCallback done!")
	}
}

func ForceGlobalDelegation(rk referralK.Keeper, bk bank.Keeper, dk delegatingK.Keeper, sk scheduleK.Keeper, bKey, dKey sdk.StoreKey, cdc codec.BinaryMarshaler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting ForceGlobalDelegation ...")
		defer logger.Info("... ForceGlobalDelegation done!")

		dMain := sdk.NewInt(0)
		dRevoke := sdk.NewInt(0)

		oneDay := sk.OneDay(ctx)
		t := time.Unix(0, plan.Time.UnixNano()).Add(oneDay)

		q := util.FractionZero()
		deltaT := -1 * time.Second
		deltaQ := util.NewFraction(-deltaT.Nanoseconds(), oneDay.Nanoseconds()).Reduce()

		bStore := ctx.KVStore(bKey)
		key := make([]byte, len(bankT.BalancesPrefix)+AddrLen)
		copy(key, bankT.BalancesPrefix)

		rk.Iterate(ctx, func(bech32 string, r *referralT.Info) (changed, _ bool) {
			if r.Banished {
				return
			}

			empty := r.Coins[0].Equal(r.Delegated[0])
			for i := 0; i <= 10; i++ {
				if !r.Coins[i].Equal(r.Delegated[i]) {
					r.Delegated[i] = r.Coins[i]
					changed = true
				}
			}
			if empty {
				return
			}

			acc, err := sdk.AccAddressFromBech32(bech32)
			if err != nil {
				logger.Error("Cannot parse account address", "acc", bech32, "err", err)
				return false, false
			}

			balance := bk.GetBalance(ctx, acc)

			mainBal := balance.AmountOf(util.ConfigMainDenom)
			if !mainBal.IsZero() {
				balance = balance.
					Sub(sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, mainBal))).
					Add(sdk.NewCoin(util.ConfigDelegatedDenom, mainBal))
			}

			revokeBal := balance.AmountOf(util.ConfigRevokingDenom)
			if !revokeBal.IsZero() {
				balance = balance.
					Sub(sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, revokeBal))).
					Add(sdk.NewCoin(util.ConfigDelegatedDenom, revokeBal))
			}

			di := dk.Get(ctx, acc)
			diChanged := false

			if di == nil {
				di = &delegatingT.Record{}
			}

			if len(di.Requests) != 0 {
				for _, r := range di.Requests {
					sk.Delete(ctx, r.Time, delegatingT.RevokeHookName, acc)
				}
				di.Requests = nil
				diChanged = true
			}
			if di.NextAccrue != nil {
				missedPart := util.NewFraction(mainBal.Int64()+revokeBal.Int64(), balance.AmountOf(util.ConfigDelegatedDenom).Int64()).Mul(util.NewFraction(plan.Time.Sub(di.NextAccrue.Add(-oneDay)).Nanoseconds(), oneDay.Nanoseconds()))
				di.MissedPart = &missedPart
				diChanged = true
			} else if balance.AmountOf(util.ConfigDelegatedDenom).Int64() > bk.GetParams(ctx).DustDelegation {
				t = t.Add(deltaT)
				q = q.Add(deltaQ)

				di.NextAccrue = &t
				di.MissedPart = &q
				diChanged = true
				sk.ScheduleTask(ctx, t, delegatingT.AccrueHookName, acc)

				if r.BanishmentAt != nil {
					sk.Delete(ctx, *r.BanishmentAt, referralK.BanishHookName, []byte(bech32))
					r.BanishmentAt = nil
					changed = true
				}
			}
			if diChanged {
				if di.IsEmpty() {
					ctx.KVStore(dKey).Delete(acc)
				} else {
					ctx.KVStore(dKey).Set(acc, cdc.MustMarshalBinaryBare(di))
				}
			}

			copy(key[len(bankT.BalancesPrefix):], acc.Bytes())
			bStore.Set(key, cdc.MustMarshalBinaryBare(&bankT.Balance{Coins: balance}))

			dMain = dMain.Add(mainBal)
			dRevoke = dRevoke.Add(revokeBal)
			return
		})

		supply := bk.GetSupply(ctx)
		supply.Deflate(sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, dMain),
			sdk.NewCoin(util.ConfigRevokingDenom, dRevoke),
		))
		supply.Inflate(sdk.NewCoins(
			sdk.NewCoin(util.ConfigDelegatedDenom, dMain.Add(dRevoke)),
		))
		bk.SetSupply(ctx, supply)

		// Referral statuses must be refreshed after this
	}
}

func RefreshReferralStatuses(rk referralK.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RefreshReferralStatuses ...")
		defer logger.Info("... RefreshReferralStatuses done!")

		rk.Iterate(ctx, func(_ string, _ *referralT.Info) (changed, checkForStatusUpdate bool) {
			return false, true
		})
	}
}

func UnbanishAccountsWithDelegation(bk bank.Keeper, sk scheduleK.Keeper, cdc codec.BinaryMarshaler, rKey sdk.StoreKey) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting UnbanishAccountsWithDelegation ...")

		store := cachekv.NewStore(ctx.KVStore(rKey))
		get := func(acc string) referralT.Info {
			bz := store.Get([]byte(acc))
			if bz == nil {
				panic("not found")
			}
			var r referralT.Info
			cdc.MustUnmarshalBinaryBare(bz, &r)
			return r
		}
		set := func(acc string, value referralT.Info) {
			store.Set([]byte(acc), cdc.MustMarshalBinaryBare(&value))
		}

		it := store.Iterator(nil, nil)
		defer func() {
			if e := recover(); e != nil {
				logger.Error("Panic during upgrade", "err", e)
			}
			if it != nil {
				_ = it.Close()
			}
			logger.Info("... UnbanishAccountsWithDelegation done!")
		}()

		ddt := bk.GetParams(ctx).DustDelegation
		logger.Debug(fmt.Sprintf("... dust delegation threshold = %d", ddt))

		for ; it.Valid(); it.Next() {
			acc := string(it.Key())
			r := get(acc)

			if !r.Banished {
				continue
			}
			addr, err := sdk.AccAddressFromBech32(acc)
			if err != nil {
				panic(errors.Wrap(err, "cannot parse address"))
			}

			if d := bk.GetBalance(ctx, addr).AmountOf(util.ConfigDelegatedDenom).Int64(); d > ddt {
				logger.Info("... unbanishing", "acc", acc, "delegation", d)

				var parent string
				var pi referralT.Info
				for parent = r.Referrer; parent != ""; parent = pi.Referrer {
					pi := get(parent)
					if !pi.RegistrationClosed(ctx, sk) {
						break
					}
				}

				r.Referrer = parent
				r.Banished = false
				r.BanishmentAt = nil
				r.CompressionAt = nil
				r.Status = referralT.STATUS_LUCKY

				set(acc, r)

				empty := true
				for i := 0; i <= 10; i++ {
					if !r.Coins[i].IsZero() || !r.Delegated[i].IsZero() {
						empty = false
						break
					}
				}

				if empty {
					logger.Debug("... ... account is empty", "acc", acc, "parent", parent)
					info := get(parent)
					info.Referrals = append(info.Referrals, acc)
					set(parent, info)
				} else {
					a := parent
					for i := 1; i <= 10; i++ {
						if a == "" {
							break
						}
						info := get(a)
						if i == 1 {
							info.Referrals = append(info.Referrals, acc)
						}
						for j := i; j <= 10; j++ {
							info.Coins[j] = info.Coins[j].Add(r.Coins[j-i])
							info.Delegated[j] = info.Delegated[j].Add(r.Delegated[j-i])
						}
						logger.Debug("... ... ancestor affected", "acc", acc, "anc", a, "lvl", i, "d", "+")
						set(a, info)
						a = info.Referrer
					}
				}
			} else if r.CompressionAt != nil {
				logger.Info("... cleaning CompressionAt", "acc", acc, "was", r.CompressionAt.String())
				r.CompressionAt = nil
				set(acc, r)
			}
		}
		_ = it.Close()
		it = nil
		store.Write()
		// RefreshReferralStatuses must be called after this.
	}
}

func TransferFromTheBanished(sk scheduleK.Keeper, cdc codec.BinaryMarshaler, rKey sdk.StoreKey) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting TransferFromTheBanished ...")

		store := cachekv.NewStore(ctx.KVStore(rKey))
		get := func(acc string) referralT.Info {
			bz := store.Get([]byte(acc))
			if bz == nil {
				panic("not found")
			}
			var r referralT.Info
			cdc.MustUnmarshalBinaryBare(bz, &r)
			return r
		}
		set := func(acc string, value referralT.Info) {
			store.Set([]byte(acc), cdc.MustMarshalBinaryBare(&value))
		}

		it := store.Iterator(nil, nil)
		defer func() {
			if e := recover(); e != nil {
				logger.Error("Panic during upgrade", "err", e)
			}
			if it != nil {
				_ = it.Close()
			}
			logger.Info("... TransferFromTheBanished done!")
		}()

		for ; it.Valid(); it.Next() {
			acc := string(it.Key())

			var r referralT.Info
			cdc.MustUnmarshalBinaryBare(it.Value(), &r)
			if r.Banished || r.Referrer == "" {
				continue
			}

			parent := r.Referrer
			pi := get(parent)
			if !pi.Banished {
				continue
			}

			logger.Info("... parent is banished, moving account up ...", "acc", acc, "parent", parent)

			for {
				parent = pi.Referrer
				if parent == "" {
					break
				}
				pi = get(parent)
				if !pi.RegistrationClosed(ctx, sk) {
					break
				}
			}

			empty := true
			for i := 0; i <= 10; i++ {
				if !r.Coins[i].IsZero() || !r.Delegated[i].IsZero() {
					empty = false
					break
				}
			}

			if empty {
				logger.Debug("... ... account is empty", "acc", acc, "former_parent", r.Referrer, "new_parent", parent)
				info := get(r.Referrer)
				util.RemoveStringPreserveOrder(&info.Referrals, acc)
				set(r.Referrer, info)

				info = get(parent)
				info.Referrals = append(info.Referrals, acc)
				set(parent, info)
			} else {
				a := r.Referrer
				for i := 1; i <= 10; i++ {
					if a == "" {
						break
					}
					info := get(a)
					if i == 1 {
						util.RemoveStringPreserveOrder(&info.Referrals, acc)
					}
					for j := i; j <= 10; j++ {
						info.Coins[j] = info.Coins[j].Sub(r.Coins[j-i])
						info.Delegated[j] = info.Delegated[j].Sub(r.Delegated[j-i])
					}
					logger.Debug("... ... ancestor affected", "acc", acc, "anc", a, "lvl", i, "d", "-")
					set(a, info)
					a = info.Referrer
				}

				a = parent
				for i := 1; i <= 10; i++ {
					if a == "" {
						break
					}
					info := get(a)
					if i == 1 {
						info.Referrals = append(info.Referrals, acc)
					}
					for j := i; j <= 10; j++ {
						info.Coins[j] = info.Coins[j].Add(r.Coins[j-i])
						info.Delegated[j] = info.Delegated[j].Add(r.Delegated[j-i])
					}
					logger.Debug("... ... ancestor affected", "acc", acc, "anc", a, "lvl", i, "d", "+")
					set(a, info)
					a = info.Referrer
				}
			}

			r.Referrer = parent
			set(acc, r)
			logger.Info("... ... relocated", "acc", acc, "parent", parent)
		}

		_ = it.Close()
		it = nil
		store.Write()
		// RefreshReferralStatuses must be called after this.
	}
}

func InitValidatorBonusParam() upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Skipping InitValidatorBonusParam")
	}
}

func InitValidatorParam(k delegatingK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitValidatorParam ...")

		var pz delegatingT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, delegatingT.KeyValidator) {
				pz.Validator = delegatingT.DefaultValidator
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitValidatorParam done!", "params", pz)
	}
}

func InitTransactionFeeParam(k bank.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitTransactionFeeParam ...")

		var pz bankT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, bankT.ParamStoreKeyTransactionFee) {
				pz.TransactionFee = bankT.DefaultTransactionFee
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitTransactionFeeParam done!", "params", pz)
	}
}

func RemovePromoBonuses(k referralK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RemovePromoBonuses ...")

		var pz referralT.Params
		for _, pair := range pz.ParamSetPairs() {
			paramspace.Get(ctx, pair.Key, pair.Value)

			if bytes.Equal(pair.Key, referralT.KeySubscriptionAward) {
				pz.SubscriptionAward.Company = pz.SubscriptionAward.Company.Add(util.Percent(5))
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... RemovePromoBonuses done!", "params", pz)
	}
}

func RemoveStatusBonuses(k referralK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RemoveStatusBonuses ...")

		var pz referralT.Params
		for _, pair := range pz.ParamSetPairs() {
			paramspace.Get(ctx, pair.Key, pair.Value)

			if bytes.Equal(pair.Key, referralT.KeySubscriptionAward) {
				pz.SubscriptionAward.Company = pz.SubscriptionAward.Company.Add(util.Percent(5))
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... RemoveStatusBonuses done!", "params", pz)
	}
}

func RemoveLeaderBonuses(k referralK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RemoveLeaderBonuses ...")

		var pz referralT.Params
		for _, pair := range pz.ParamSetPairs() {
			paramspace.Get(ctx, pair.Key, pair.Value)

			if bytes.Equal(pair.Key, referralT.KeySubscriptionAward) {
				pz.SubscriptionAward.Company = pz.SubscriptionAward.Company.Add(util.Percent(5))
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... RemoveLeaderBonuses done!", "params", pz)
	}
}

func InitBurnOnRevokeParam(k delegatingK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitBurnOnRevokeParam ...")

		var pz delegatingT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, delegatingT.KeyBurnOnRevoke) {
				pz.BurnOnRevoke = delegatingT.DefaultBurnOnRevoke
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitBurnOnRevokeParam done!", "params", pz)
	}
}

func UpdateStatusDowngradeTasks(sk scheduleK.Keeper, rKey, sKey sdk.StoreKey, cdc codec.BinaryMarshaler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting UpdateStatusDowngradeTasks ...")

		{
			var (
				ms  = ctx.MultiStore().CacheMultiStore()
				ctx = ctx.WithMultiStore(ms)

				sStore = ctx.KVStore(sKey)
				sIt    = sStore.Iterator(nil, nil)

				rStore = ctx.KVStore(rKey)

				newDowngradeTime = ctx.BlockTime().Add(2 * sk.OneDay(ctx))
			)
			logger.Info("... fixing schedule ...")
			for ; sIt.Valid(); sIt.Next() {
				var sch scheduleT.Schedule
				cdc.MustUnmarshalBinaryBare(sIt.Value(), &sch)

				changed := false
				tasks := make([]scheduleT.Task, 0, len(sch.Tasks))

				for _, task := range sch.Tasks {
					taskUpdated := false
					if task.HandlerName == referralK.StatusDowngradeHookName {
						logger.Debug("find hook", "handlerName", task.HandlerName, "time", task.Time, "data", task.Data)
						if task.Time.After(newDowngradeTime) {
							task.Time = newDowngradeTime
							taskUpdated = true
							logger.Debug("update hook", "handlerName", task.HandlerName, "time", task.Time, "data", task.Data)
						}
					}
					tasks = append(tasks, task)
					if taskUpdated {
						changed = true
					}
				}

				if changed {
					if len(tasks) == 0 {
						sStore.Delete(sIt.Key())
					} else {
						sch.Tasks = tasks
						sStore.Set(sIt.Key(), cdc.MustMarshalBinaryBare(&sch))
					}
				}
			}
			sIt.Close()

			var (
				rIt = rStore.Iterator(nil, nil)
			)
			logger.Info("... fixing referral ...")
			for ; rIt.Valid(); rIt.Next() {
				acc := string(rIt.Key())

				var r referralT.Info
				cdc.MustUnmarshalBinaryBare(rIt.Value(), &r)
				if r.StatusDowngradeAt != nil {
					logger.Debug("find referral StatusDowngradeAt", "acc", acc, "statusDowngradeAt", r.StatusDowngradeAt.String())
					if r.StatusDowngradeAt.After(newDowngradeTime) {
						r.StatusDowngradeAt = &newDowngradeTime
						rStore.Set(rIt.Key(), cdc.MustMarshalBinaryBare(&r))
						logger.Debug("update referral StatusDowngradeAt", "acc", acc, "statusDowngradeAt", r.StatusDowngradeAt.String())
					}
				}
			}
			rIt.Close()

			logger.Debug("... persisting multistore ...")
			ms.Write()
		}
		logger.Info("... UpdateStatusDowngradeTasks done!")
	}
}

func InitMaxTransactionFeeParam(k bank.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitMaxTransactionFeeParam ...")

		var pz bankT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, bankT.ParamStoreKeyMaxTransactionFee) {
				pz.MaxTransactionFee = bankT.DefaultMaxTransactionFee
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitMaxTransactionFeeParam done!", "params", pz)
	}
}

func FixStatusDowngradeTasks(sk scheduleK.Keeper, sKey sdk.StoreKey, cdc codec.BinaryMarshaler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting UpdateStatusDowngradeTasks ...")

		{
			var (
				ms  = ctx.MultiStore().CacheMultiStore()
				ctx = ctx.WithMultiStore(ms)

				sStore = ctx.KVStore(sKey)
				sIt    = sStore.Iterator(nil, nil)

				blockTime = ctx.BlockTime()

				incorrectTasks = make([]scheduleT.Task, 0, 512)
			)
			logger.Info("... fixing schedule ...")
			for ; sIt.Valid(); sIt.Next() {
				var sch scheduleT.Schedule
				cdc.MustUnmarshalBinaryBare(sIt.Value(), &sch)

				changed := false
				tasks := make([]scheduleT.Task, 0, len(sch.Tasks))

				for _, task := range sch.Tasks {
					taskDeleted := false
					if task.HandlerName == referralK.StatusDowngradeHookName {
						logger.Debug("find hook", "handlerName", task.HandlerName, "t", task.Time.String(), "data", task.Data)
						if !bytes.Equal(sIt.Key(), scheduleK.Key(task.Time)) && task.Time.Before(blockTime) {
							logger.Debug("need fix hook", "handlerName", task.HandlerName, "t", task.Time.String(), "data", task.Data)
							taskDeleted = true
						}
					}
					if taskDeleted {
						incorrectTasks = append(incorrectTasks, task)
						changed = true
					} else {
						tasks = append(tasks, task)
					}
				}

				if changed {
					if len(tasks) == 0 {
						sStore.Delete(sIt.Key())
					} else {
						sch.Tasks = tasks
						sStore.Set(sIt.Key(), cdc.MustMarshalBinaryBare(&sch))
					}
				}
			}
			sIt.Close()

			var (
				shiftInterval    = time.Duration(sk.GetParams(ctx).DayNanos * 15 / (24 * 60 * 60))
				newDowngradeTime = ctx.BlockTime()
			)
			for _, task := range incorrectTasks {
				newDowngradeTime = newDowngradeTime.Add(shiftInterval)
				fixKey := scheduleK.Key(newDowngradeTime)
				var fixSch scheduleT.Schedule
				if err := cdc.UnmarshalBinaryBare(fixKey, &fixSch); err != nil {
					logger.Debug("not found key by", "t", newDowngradeTime.String())
				}
				fixSch.Tasks = append(fixSch.Tasks, task)
				sStore.Set(fixKey, cdc.MustMarshalBinaryBare(&fixSch))
				logger.Debug("relocate hook", "handlerName", task.HandlerName, "t", task.Time.String(), "data", task.Data)
			}

			logger.Debug("... persisting multistore ...")
			ms.Write()
		}
		logger.Info("... UpdateStatusDowngradeTasks done!")
	}
}

func ScheduleMissingBanishmentAndRefreshReferralStatuses(rk referralK.Keeper, bk bank.Keeper, sk scheduleK.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting ScheduleMissingBanishmentAndRefreshReferralStatuses ...")
		defer logger.Info("... ScheduleMissingBanishmentAndRefreshReferralStatuses done!")

		var (
			dd = bk.GetParams(ctx).DustDelegation

			blockTime         = ctx.BlockTime()
			shiftInterval     = time.Duration(sk.GetParams(ctx).DayNanos * 15 / (24 * 60 * 60))
			newBanishmentTime = blockTime.Add(time.Duration(sk.GetParams(ctx).DayNanos / 24))
		)

		rk.Iterate(ctx, func(bech32 string, r *referralT.Info) (changed, checkForStatusUpdate bool) {
			changed = false
			checkForStatusUpdate = true

			acc, err := sdk.AccAddressFromBech32(bech32)
			if err != nil {
				logger.Error("Cannot parse account address", "acc", bech32, "err", err)
				return
			}

			balance := bk.GetBalance(ctx, acc)

			if !r.Active {
				if balance.AmountOf(util.ConfigDelegatedDenom).Int64() <= dd {
					if !r.Banished && r.BanishmentAt == nil && (r.CompressionAt == nil || blockTime.After(*r.CompressionAt)) {
						newBanishmentTime = newBanishmentTime.Add(shiftInterval)
						logger.Debug("add missing banishment", "acc", bech32, "t", newBanishmentTime.String())
						sk.ScheduleTask(ctx, newBanishmentTime, referralK.BanishHookName, []byte(bech32))
						r.BanishmentAt = &newBanishmentTime
						changed = true
					}
				}
			}

			return
		})
	}
}

func InitTransactionFeeSplitRatiosAndCompanyAccountParams(k bank.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitTransactionFeeSplitRatiosAndCompanyAccountParams ...")

		var pz bankT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, bankT.ParamStoreKeyTransactionFeeSplitRatios) {
				pz.TransactionFeeSplitRatios = bankT.DefaultTransactionFeeSplitRatios
			} else if bytes.Equal(pair.Key, bankT.ParamStoreKeyCompanyAccount) {
				pz.CompanyAccount = "artr1d3paqmusp39t2yhx4ju4vm50pfjmddfkwnn22p"
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitTransactionFeeSplitRatiosAndCompanyAccountParams done!", "params", pz)
	}
}

func InitAccruePercentageRangesAndValidatorBonusParams(k delegatingK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitAccruePercentageRangesAndValidatorBonusParams ...")

		var pz delegatingT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, delegatingT.KeyAccruePercentageRanges) {
				pz.AccruePercentageRanges = []delegatingT.PercentageRange{
					{Start: 0, Percent: util.Percent(pz.Percentage.Minimal)},
					{Start: 1_000_000000, Percent: util.Percent(pz.Percentage.ThousandPlus)},
					{Start: 10_000_000000, Percent: util.Percent(pz.Percentage.TenKPlus)},
					{Start: 100_000_000000, Percent: util.Percent(pz.Percentage.HundredKPlus)},
				}
			} else if bytes.Equal(pair.Key, delegatingT.KeyValidatorBonus) {
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		pz.ValidatorBonus = pz.Validator.Sub(util.Percent(pz.Percentage.HundredKPlus))
		if pz.ValidatorBonus.IsNegative() {
			pz.ValidatorBonus = util.FractionZero()
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitAccruePercentageRangesAndValidatorBonusParams done!", "params", pz)
	}
}

func InitBlockedSendersParam(k bank.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitBlockedSendersParam ...")

		var pz bankT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, bankT.ParamStoreKeyBlockedSenders) {
				pz.SetBlockedSenders(bankT.DefaultBlockedSenders)
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitBlockedSendersParam done!", "params", pz)
	}
}

func CleanEarningStore(storeKey sdk.StoreKey) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting CleanEarningStore ...")

		store := ctx.KVStore(storeKey)
		var keys [][]byte
		it := store.Iterator(nil, nil)
		for ; it.Valid(); it.Next() {
			keys = append(keys, it.Key())
		}
		it.Close()
		for _, key := range keys {
			store.Delete(key)
		}

		logger.Info("... CleanEarningStore done!")
	}
}

func InitSubscriptionVpnStorageBonusesParams(k delegatingK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitSubscriptionVpnStorageBonusesParams ...")

		var pz delegatingT.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, delegatingT.KeySubscriptionBonus) {
				pz.SubscriptionBonus = delegatingT.DefaultSubscriptionBonus
			} else if bytes.Equal(pair.Key, delegatingT.KeyVpnBonus) {
				pz.VpnBonus = delegatingT.DefaultVpnBonus
			} else if bytes.Equal(pair.Key, delegatingT.KeyStorageBonus) {
				pz.StorageBonus = delegatingT.DefaultStorageBonus
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)
		logger.Info("... InitSubscriptionVpnStorageBonusesParams done!", "params", pz)
	}
}
