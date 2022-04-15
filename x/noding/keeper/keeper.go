package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/noding/types"
)

// Keeper of the noding store
type Keeper struct {
	dataStoreKey     sdk.StoreKey
	indexStoreKey    sdk.StoreKey
	cdc              codec.BinaryMarshaler
	referralKeeper   types.ReferralKeeper
	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	paramspace       types.ParamSubspace
	feeCollectorName string
}

// NewKeeper creates a noding keeper
func NewKeeper(
	cdc codec.BinaryMarshaler,
	dataKey sdk.StoreKey,
	indexKey sdk.StoreKey,
	referralKeeper types.ReferralKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	paramspace types.ParamSubspace,
	feeCollectorName string,
) Keeper {
	keeper := Keeper{
		dataStoreKey:     dataKey,
		indexStoreKey:    indexKey,
		cdc:              cdc,
		referralKeeper:   referralKeeper,
		accountKeeper:    accountKeeper,
		bankKeeper:       bankKeeper,
		paramspace:       paramspace.WithKeyTable(types.ParamKeyTable()),
		feeCollectorName: feeCollectorName,
	}
	return keeper
}

var IdxPrefixNodeOperator = []byte{0x01}
var IdxPrefixBlockProposer = []byte{0x02}
var IdxPrefixLotteryQueue = []byte{0x03}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) IsQualified(ctx sdk.Context, accAddr sdk.AccAddress) (result bool, delegation sdk.Int, reason types.Reason, err error) {
	delegation, err = k.referralKeeper.GetDelegatedInNetwork(ctx, accAddr.String(), 10)
	if err != nil {
		return
	}

	// Check if it's staff
	if k.has(ctx, accAddr) {
		var d types.Info
		d, err = k.Get(ctx, accAddr)
		if err != nil {
			return
		}
		if d.Staff {
			result = true
		}
	}

	// Check minimal status
	status, err := k.referralKeeper.GetStatus(ctx, accAddr.String())
	if err != nil {
		return
	}
	if !result && status < k.GetParams(ctx).MinStatus {
		reason = types.REASON_NOT_ENOUGH_STATUS
		return
	}

	// 10k ARTR delegated
	if !result && delegation.Int64() < 10_000_000000 {
		reason = types.REASON_NOT_ENOUGH_STAKE
		return
	}

	result = true
	return
}

// IsValidator returns true if an account presents in the validator pool
// (i.e. if it can be potentially chosen for block signing).
func (k Keeper) IsValidator(ctx sdk.Context, accAddr sdk.AccAddress) (bool, error) {
	if !k.has(ctx, accAddr) {
		return false, nil
	}
	record, err := k.Get(ctx, accAddr)
	if err != nil {
		return false, err
	}
	return record.IsActive(), nil
}

func (k Keeper) IsActiveValidator(ctx sdk.Context, accAddr sdk.AccAddress) (bool, error) {
	if !k.has(ctx, accAddr) {
		return false, nil
	}
	record, err := k.Get(ctx, accAddr)
	if err != nil {
		return false, err
	}
	if !record.IsActive() {
		return false, nil
	}
	pz := k.GetParams(ctx)
	if ctx.BlockHeight() - record.UnjailAt < 6 * int64(pz.UnjailAfter) {
		return false, nil
	}
	return true, nil
}

func (k Keeper) IsBanned(ctx sdk.Context, accAddr sdk.AccAddress) (bool, error) {
	if !k.has(ctx, accAddr) {
		return false, nil
	}
	record, err := k.Get(ctx, accAddr)
	if err != nil {
		return false, err
	}
	return record.BannedForLife, nil
}

func (k Keeper) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (result sdk.AccAddress, found bool, active bool, err error) {
	result, found = k.getNodeOperatorFromIndex(ctx, consAddr)
	if found {
		var data types.Info
		data, err = k.Get(ctx, result)
		if err == nil && data.IsActive() {
			active = consAddr.Equals(consAddressFromCryptoBubKey(cryptoPubKeyFromBech32(data.PubKey)))
		}
	}
	return
}

func (k Keeper) SwitchOn(ctx sdk.Context, accAddr sdk.AccAddress, key crypto.PubKey) error {
	isBanned, err := k.IsBanned(ctx, accAddr)
	if err != nil {
		return errors.Wrap(err, "cannot check for ban")
	}
	if isBanned {
		return types.ErrBannedForLifetime
	}

	if k.has(ctx, accAddr) {
		data, err := k.Get(ctx, accAddr)
		if err != nil {
			return sdkerrors.Wrapf(err, "cannot get data for %s", accAddr.String())
		}
		if data.Status {
			return types.ErrAlreadyOn
		}
	}

	consAddr := sdk.GetConsAddress(key)
	_, found, active, err := k.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		k.Logger(ctx).Error("couldn't Get validator by consensus address", "consAddr", consAddr)
		return err
	}
	if found && active {
		k.Logger(ctx).Error("validator with same public key already exists", "pubKey", key)
		return types.ErrPubkeyBusy
	}

	isQualified, delegation, reason, err := k.IsQualified(ctx, accAddr)
	if err != nil {
		k.Logger(ctx).Error("k.IsQualified failed", "account", accAddr, "error", err)
		return err
	}
	if !isQualified {
		k.Logger(ctx).Error("account is not qualified for noding", "account", accAddr, "reason", reason)
		return types.ErrNotQualified
	}

	if k.has(ctx, accAddr) {
		err = k.update(ctx, accAddr, func(d *types.Info) (save bool) {
			d.Status = true
			d.PubKey = bech32FromCryptoPubKey(key)
			d.UpdateScore(delegation.Int64())
			return true
		})
	} else {
		err = k.set(ctx, accAddr, *types.NewInfo(bech32FromCryptoPubKey(key), delegation.Int64()))
	}
	if err != nil {
		return err
	}

	k.addToIndex(ctx, nodeOperatorIdxKey(consAddressFromCryptoBubKey(key)), accAddr.Bytes())
	return nil
}

func (k Keeper) SwitchOff(ctx sdk.Context, accAddr sdk.AccAddress) error {
	err := k.update(ctx, accAddr, func(d *types.Info) (save bool) {
		if !d.Status {
			return false
		}

		d.Status = false
		if d.LotteryNo != 0 {
			if err := k.lotteryExclude(ctx, d); err != nil {
				// should never happen
				panic(err)
			}
		}
		return true
	})
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) OnStatusUpdate(ctx sdk.Context, acc sdk.AccAddress) error {
	is, err := k.IsValidator(ctx, acc)
	if err != nil {
		return err
	}
	if !is {
		return nil
	}

	var reason types.Reason
	is, _, reason, err = k.IsQualified(ctx, acc)
	if err != nil {
		return err
	}
	if is {
		return nil
	}

	k.Logger(ctx).Info("not qualified anymore, banishing from validators", "account", acc)
	err = k.SwitchOff(ctx, acc)
	if err != nil {
		return err
	}

	util.EmitEvent(ctx,
		&types.EventValidatorBanished{
			Address: acc.String(),
			Reason:  reason,
		},
	)
	return nil
}

func (k Keeper) OnStakeChanged(ctx sdk.Context, acc sdk.AccAddress) error {
	record, err := k.Get(ctx, acc)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) { return nil }
		return errors.Wrap(err, "cannot obtain data")
	}
	isV := record.IsActive()
	isQ, delegation, reason, err := k.IsQualified(ctx, acc)
	if err != nil {	return errors.Wrap(err, "cannot check if validator's qualified") }

	changed := record.UpdateScore(delegation.Int64())

	if isV && !isQ {
		k.Logger(ctx).Info("not qualified anymore, banishing from validators", "account", acc)

		record.Status = false
		if record.LotteryNo != 0 {
			if err := k.lotteryExclude(ctx, &record); err != nil {
				// should never happen
				panic(err)
			}
		}
		changed = true

		util.EmitEvent(ctx,
			&types.EventValidatorBanished{
				Address: acc.String(),
				Reason:  reason,
			},
		)
		return nil
	}

	if changed {
		if err := k.set(ctx, acc, record); err != nil {
			return errors.Wrap(err, "cannot set data")
		}
	}
	return nil
}

func (k Keeper) GatherValidatorUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	var (
		store = ctx.KVStore(k.dataStoreKey)

		result []abci.ValidatorUpdate
		active []types.InfoWithAccount
	)

	it := store.Iterator(nil, nil)
	for ; it.Valid(); it.Next() {
		var (
			addr sdk.AccAddress
			data types.Info
		)
		addr = sdk.AccAddress(it.Key())
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &data)
		if data.IsActive() {
			active = append(active, types.NewInfoWithAccount(addr, data))
			consAddress := consAddressFromCryptoBubKey(cryptoPubKeyFromBech32(data.PubKey))
			k.addNodeOperatorToIndex(ctx, consAddress, addr)
		} else {
			if data.LastPower != 0 {
				if len(data.LastPubKey) == 0 {
					panic("non-zero LastPower is impossible without LastPubKey")
				}

				result = append(result, abci.ValidatorUpdate{
					PubKey: abciPubKeyFromBech32(data.LastPubKey),
					Power:  0,
				})
				if err := k.update(ctx, addr, func(d *types.Info) (save bool) {
					d.LastPower = 0
					d.LastPubKey = ""
					return true
				}); err != nil {
					defer it.Close()
					return nil, err
				}
			}
		}
	}
	it.Close()

	var (
		params             = k.GetParams(ctx)
		maxTopValidators   = int(params.MaxValidators)
		maxLuckyValidators = int(params.LotteryValidators)
		totalMaxValidators = maxTopValidators + maxLuckyValidators

		i, n1, n2 int
	)

	if len(active) > maxTopValidators {
		n1 = maxTopValidators
		if len(active) > totalMaxValidators {
			n2 = maxLuckyValidators
		} else {
			n2 = len(active) - n1
		}
	} else {
		n1 = len(active)
		n2 = 0
	}
	sort.Slice(active, func(i, j int) bool {
		if active[i].Score != active[j].Score {
			return active[i].Score > active[j].Score
		}
		return active[i].OkBlocksInRow > active[j].OkBlocksInRow
	})
	vpg := NewVotingPowerGenerator(k, ctx)

	for i = 0; i < n1; i++ {
		data := active[i]
		if len(data.PubKey) == 0 {
			panic("validator cannot be active without PubKey")
		}

		power := vpg.GetVotingPower(data.Score)
		updated := data.LastPower != power
		if data.PubKey != data.LastPubKey {
			if len(data.LastPubKey) != 0 {
				result = append(result, abci.ValidatorUpdate{
					PubKey: abciPubKeyFromBech32(data.LastPubKey),
					Power:  0,
				})
			}
			updated = true
		}

		if updated {
			result = append(result, abci.ValidatorUpdate{
				PubKey: abciPubKeyFromBech32(data.PubKey),
				Power:  power,
			})
		}
		if data.LotteryNo != 0 {
			if err := k.lotteryExclude(ctx, &data.Info); err != nil {
				// Should never happen, we've just checked they has been participating
				panic(err)
			}
			updated = true
		}
		if updated {
			if err := k.update(ctx, data.Account, func(d *types.Info) (save bool) {
				d.LastPower = power
				d.LastPubKey = d.PubKey
				d.LotteryNo = data.LotteryNo
				return true
			}); err != nil {
				return nil, err
			}
		}
	}
	maxLotNo := k.lotteryLastNo(ctx, n2)
	for ; i < len(active); i++ {
		data := active[i]
		if data.LotteryNo != 0 && data.LotteryNo <= maxLotNo {
			power := vpg.GetVotingPower(data.Score)
			if data.PubKey != data.LastPubKey {
				if len(data.LastPubKey) != 0 {
					result = append(result, abci.ValidatorUpdate{
						PubKey: abciPubKeyFromBech32(data.LastPubKey),
						Power:  0,
					})
				}
			} else if data.LastPower == power {
				continue
			}

			result = append(result, abci.ValidatorUpdate{
				PubKey: abciPubKeyFromBech32(data.PubKey),
				Power:  power,
			})
			if err := k.update(ctx, data.Account, func(d *types.Info) (save bool) {
				d.LastPower = power
				d.LastPubKey = d.PubKey
				return true
			}); err != nil {
				return nil, err
			}
		} else {
			updated := false
			if data.LastPower != 0 {
				result = append(result, abci.ValidatorUpdate{
					PubKey: abciPubKeyFromBech32(data.LastPubKey),
					Power:  0,
				})
				updated = true
			}
			if data.LotteryNo == 0 {
				if err := k.lotteryAddNew(ctx, data.Account, &data.Info); err != nil {
					// Should never happen, we've just checked that the account hasn't been participating yet.
					panic(err)
				}
				updated = true
			}
			if updated {
				if err := k.update(ctx, data.Account, func(d *types.Info) (save bool) {
					d.LastPower = 0
					d.LastPubKey = ""
					d.LotteryNo = data.LotteryNo
					return true
				}); err != nil {
					return nil, err
				}
			}
		}
	}

	// Just in case an operator switches node off and another one switches it on immediately
	unique := make([]abci.ValidatorUpdate, 0, len(result))
	for _, x := range result {
		found := false
		for j, y := range unique {
			if x.PubKey == y.PubKey {
				unique[j] = x
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, x)
		}
	}

	return unique, nil
}

// GetActiveValidators - returns all potential (i.e. switched on and not jailed) validators. Not just a top N, that is
// chosen for tendermint consensus, all of them.
func (k Keeper) GetActiveValidators(ctx sdk.Context) ([]types.Validator, error) {
	var result []types.Validator
	store := ctx.KVStore(k.dataStoreKey)
	proposed := k.GetBlocksProposedByAll(ctx)

	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var value types.Info
		if err := proto.Unmarshal(it.Value(), &value); err != nil {
			panic(errors.Wrap(err, "cannot unmarshal info"))
		}
		if !value.IsActive() {
			continue
		}
		addr := sdk.AccAddress(it.Key())
		result = append(result, types.GenesisValidatorFromD(addr, value, proposed[addr.String()]))
	}

	return result, nil
}

// GetActiveValidatorList is just like GetActiveValidators but returns sccount addresses only without any detail.
func (k Keeper) GetActiveValidatorList(ctx sdk.Context) ([]sdk.AccAddress, error) {
	var result []sdk.AccAddress
	store := ctx.KVStore(k.dataStoreKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var value types.Info
		if err := proto.Unmarshal(it.Value(), &value); err != nil {
			panic(errors.Wrap(err, "cannot unmarshal info"))
		}
		if !value.IsActive() {
			continue
		}
		addr := sdk.AccAddress(it.Key())
		result = append(result, addr)
	}

	return result, nil
}

func (k Keeper) GetNonActiveValidators(ctx sdk.Context) ([]types.Validator, error) {
	var result []types.Validator
	store := ctx.KVStore(k.dataStoreKey)
	proposed := k.GetBlocksProposedByAll(ctx)

	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var value types.Info
		if err := proto.Unmarshal(it.Value(), &value); err != nil {
			panic(errors.Wrap(err, "cannot unmarshal info"))
		}
		if value.IsActive() {
			continue
		}
		addr := sdk.AccAddress(it.Key())
		result = append(result, types.GenesisValidatorFromD(addr, value, proposed[addr.String()]))
	}

	return result, nil
}

func (k Keeper) SetActiveValidators(ctx sdk.Context, validators []types.Validator) error {
	for _, v := range validators {
		pubkey := cryptoPubKeyFromBech32(v.PubKey)

		acc := v.GetAccount()
		stake, err := k.referralKeeper.GetDelegatedInNetwork(ctx, acc.String(), 10)
		if err != nil { return errors.Wrap(err, "cannot obtain stake") }
		if err := k.set(ctx, acc, v.ToInfo(stake.Int64())); err != nil {
			return errors.Wrap(err, "cannot set data")
		}
		for _, h := range v.ProposedBlocks {
			k.addProposerToIndex(ctx, int64(h), acc)
		}
		if err := k.SwitchOn(ctx, acc, pubkey); err != nil {
			return errors.Wrap(err, "cannot switch on")
		}
	}
	return nil
}

func (k Keeper) SetNonActiveValidators(ctx sdk.Context, validators []types.Validator) error {
	for _, v := range validators {
		acc := v.GetAccount()
		stake, err := k.referralKeeper.GetDelegatedInNetwork(ctx, acc.String(), 10)
		if err != nil { return errors.Wrap(err, "cannot obtain stake") }
		if err := k.set(ctx, acc, v.ToInfo(stake.Int64())); err != nil {
			return err
		}
		for _, h := range v.ProposedBlocks {
			k.addProposerToIndex(ctx, int64(h), acc)
		}
	}
	return nil
}

// MarkStroke - to be called every time the validator misses a block.
func (k Keeper) MarkStroke(ctx sdk.Context, acc sdk.AccAddress) error {
	p := k.GetParams(ctx)

	return k.update(ctx, acc, func(d *types.Info) (save bool) {
		if d.Jailed {
			return false
		}

		d.Score = d.Score - d.OkBlocksInRow/100 - 1
		d.Strokes++
		d.OkBlocksInRow = 0
		d.MissedBlocksInRow++
		if d.MissedBlocksInRow >= int64(p.JailAfter) {
			d.Jailed = true
			d.UnjailAt = ctx.BlockHeight() + int64(p.UnjailAfter)
			d.JailCount++
			d.MissedBlocksInRow = 0
			if d.LotteryNo != 0 {
				if err := k.lotteryExclude(ctx, d); err != nil {
					// Should never happen
					panic(err)
				}
			}
			util.EmitEvent(ctx,
				&types.EventValidatorJailed{
					Address: acc.String(),
				},
			)
		} else {
			if d.LotteryNo != 0 {
				if err := k.lotteryDownshift(ctx, acc, d); err != nil {
					// Should never happen
					panic(err)
				}
			}
		}
		return true
	})
}

// MarkTick - to be called every time the validator signs a block successfully.
func (k Keeper) MarkTick(ctx sdk.Context, acc sdk.AccAddress) error {
	return k.update(ctx, acc, func(d *types.Info) (save bool) {
		d.MissedBlocksInRow = 0
		d.OkBlocksInRow++
		if d.OkBlocksInRow % 100 == 0 {
			d.Score++
		}
		return true
	})
}

func (k Keeper) MarkByzantine(ctx sdk.Context, acc sdk.AccAddress, evidence abci.Evidence) error {
	return k.update(ctx, acc, func(d *types.Info) (save bool) {
		d.Infractions = append(d.Infractions, evidence)
		event := types.EventByzantine{
			Address:   acc.String(),
			Evidences: d.Infractions,
		}
		if len(d.Infractions) > 1 {
			d.BannedForLife = true
			d.Status = false
			if d.LotteryNo != 0 {
				if err := k.lotteryExclude(ctx, d); err != nil {
					// Should never happen
					panic(err)
				}
			}
			event.Banned = true
		} else {
			if d.LotteryNo != 0 {
				if err := k.lotteryDownshift(ctx, acc, d); err != nil {
					// Should never happen
					panic(err)
				}
			}
			event.Banned = false
		}
		util.EmitEvent(ctx, &event)
		return true
	})
}

func (k Keeper) Unjail(ctx sdk.Context, acc sdk.AccAddress) error {
	data, err := k.Get(ctx, acc)
	if err != nil {
		return err
	}
	if !data.Jailed {
		return types.ErrNotJailed
	}
	if ctx.BlockHeight() < data.UnjailAt {
		return types.ErrJailPeriodNotOver
	}
	data.Jailed = false
	q, delegation, reason, err := k.IsQualified(ctx, acc)
	if err != nil {
		return err
	}
	if q {
		data.UpdateScore(delegation.Int64())
	} else {
		k.Logger(ctx).Info("banishing from validators", "acc", acc, "reason", reason)
		data.Status = false
		util.EmitEvent(ctx,
			&types.EventValidatorBanished{
				Address: acc.String(),
				Reason:  reason,
			},
		)
	}
	err = k.set(ctx, acc, data)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) PayProposerReward(ctx sdk.Context, acc sdk.AccAddress) (err error) {
	k.addProposerToIndex(ctx, ctx.BlockHeight()-1, acc)
	if err := k.update(ctx, acc, func(d *types.Info) (save bool) {
		d.ProposedCount++
		if d.LotteryNo != 0 {
			if err := k.lotteryDownshift(ctx, acc, d); err != nil {
				// should never happen
				panic(err)
			}
		}
		return true
	}); err != nil {
		return err
	}

	amount := k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(k.feeCollectorName))
	if amount.IsZero() {
		return nil
	}
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, acc, amount); err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetBlockProposer(ctx sdk.Context, height int64) (sdk.AccAddress, error) {
	result, found := k.getProposerFromIndex(ctx, height)
	if !found {
		return nil, types.ErrNotFound
	}
	return result, nil
}

func (k Keeper) AddToStaff(ctx sdk.Context, acc sdk.AccAddress) (err error) {
	if k.has(ctx, acc) {
		err = k.update(ctx, acc, func(d *types.Info) (save bool) {
			if d.Staff {
				return false
			}

			d.Staff = true
			return true
		})
	} else {
		err = k.set(ctx, acc, types.Info{Staff: true})
	}
	return err
}

func (k Keeper) RemoveFromStaff(ctx sdk.Context, acc sdk.AccAddress) (err error) {
	if !k.has(ctx, acc) {
		return nil
	}

	var isActive bool
	err = k.update(ctx, acc, func(d *types.Info) (save bool) {
		isActive = d.Staff && d.IsActive()

		if !d.Staff {
			return false
		}

		d.Staff = false
		return true
	})
	if err != nil {
		return err
	}

	if isActive {
		if err = k.OnStatusUpdate(ctx, acc); err != nil {
			return err
		}
		if err = k.OnStakeChanged(ctx, acc); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) GetBlocksProposedBy(ctx sdk.Context, acc sdk.AccAddress) (heights []uint64) {
	it := sdk.KVStorePrefixIterator(ctx.KVStore(k.indexStoreKey), IdxPrefixBlockProposer)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		if bytes.Equal(it.Value(), acc.Bytes()) {
			heights = append(heights, binary.BigEndian.Uint64(it.Key()[len(IdxPrefixBlockProposer):]))
		}
	}
	return heights
}

func (k Keeper) GetBlocksProposedByAll(ctx sdk.Context) (heightsByAccAddress map[string][]uint64) {
	heightsByAccAddress = make(map[string][]uint64)
	it := sdk.KVStorePrefixIterator(ctx.KVStore(k.indexStoreKey), IdxPrefixBlockProposer)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		key := sdk.AccAddress(it.Value()).String()
		height := binary.BigEndian.Uint64(it.Key()[len(IdxPrefixBlockProposer):])
		heightsByAccAddress[key] = append(heightsByAccAddress[key], height)
	}
	return heightsByAccAddress
}

func (k Keeper) GeneralAmnesty(ctx sdk.Context) {
	store := ctx.KVStore(k.dataStoreKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var item types.Info
		if err := proto.Unmarshal(it.Value(), &item); err != nil {
			panic(errors.Wrap(err, "cannot unmarshal info"))
		}
		if item.Strokes == 0 && item.JailCount == 0 {
			continue
		}
		item.Score += item.Strokes
		item.Strokes = 0
		item.JailCount = 0
		if bz, err := proto.Marshal(&item); err != nil {
			panic(errors.Wrap(err, "cannot marshal info"))
		} else {
			store.Set(it.Key(), bz)
		}
	}
}

func (k Keeper) GetValidatorState(ctx sdk.Context, acc sdk.AccAddress) types.ValidatorState {
	data, err := k.Get(ctx, acc)
	if err != nil {
		return types.VALIDATOR_STATE_OFF
	}
	if data.BannedForLife {
		return types.VALIDATOR_STATE_BAN
	}
	if data.Jailed {
		return types.VALIDATOR_STATE_JAIL
	}
	if !data.Status {
		return types.VALIDATOR_STATE_OFF
	}
	if data.LastPower == 0 {
		return types.VALIDATOR_STATE_SPARE
	}
	if data.LotteryNo == 0 {
		return types.VALIDATOR_STATE_TOP
	}
	return types.VALIDATOR_STATE_LUCKY
}

//----------------------------------------------------------------------------------

func (k Keeper) has(ctx sdk.Context, acc sdk.AccAddress) bool {
	return ctx.KVStore(k.dataStoreKey).Has(acc)
}

func (k Keeper) Get(ctx sdk.Context, acc sdk.AccAddress) (types.Info, error) {
	store := ctx.KVStore(k.dataStoreKey)
	key := []byte(acc)
	if !store.Has(key) {
		return types.Info{}, types.ErrNotFound
	}
	var item types.Info
	err := proto.Unmarshal(store.Get(key), &item)
	return item, err
}

func (k Keeper) set(ctx sdk.Context, acc sdk.AccAddress, value types.Info) error {
	store := ctx.KVStore(k.dataStoreKey)
	keyBytes := []byte(acc)
	valueBytes, err := proto.Marshal(&value)
	if err != nil {
		return err
	}

	store.Set(keyBytes, valueBytes)
	return nil
}

func (k Keeper) update(ctx sdk.Context, acc sdk.AccAddress, callback func(d *types.Info) (save bool)) error {
	var (
		store      = ctx.KVStore(k.dataStoreKey)
		keyBytes   = []byte(acc)
		value      types.Info
	)
	if !store.Has(keyBytes) {
		return types.ErrNotFound
	}
	k.cdc.MustUnmarshalBinaryBare(store.Get(keyBytes), &value)
	if !callback(&value) {
		return nil
	}

	store.Set(keyBytes, k.cdc.MustMarshalBinaryBare(&value))
	return nil
}

func abciPubKeyFromBech32(bech32 string) tmcrypto.PublicKey {
	key, err := cryptocodec.ToTmProtoPublicKey(cryptoPubKeyFromBech32(bech32))
	if err != nil {
		panic(err)
	}
	return key
}

func cryptoPubKeyFromBech32(bech32 string) crypto.PubKey {
	return sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, bech32)
}

func bech32FromCryptoPubKey(key crypto.PubKey) string {
	return sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, key)
}

func consAddressFromCryptoBubKey(key crypto.PubKey) sdk.ConsAddress {
	return sdk.ConsAddress(key.Address().Bytes())
}

func nodeOperatorIdxKey(address sdk.ConsAddress) []byte {
	pfxLen := len(IdxPrefixNodeOperator)
	result := make([]byte, pfxLen+len(address.Bytes()))
	copy(result[:pfxLen], IdxPrefixNodeOperator)
	copy(result[pfxLen:], address.Bytes())
	return result
}

func proposerIdxKey(height int64) []byte {
	n := len(IdxPrefixBlockProposer)
	key := make([]byte, n+8)
	copy(key[:n], IdxPrefixBlockProposer)
	binary.BigEndian.PutUint64(key[n:], uint64(height))
	return key
}

func (k Keeper) addNodeOperatorToIndex(ctx sdk.Context, consAddr sdk.ConsAddress, accAddr sdk.AccAddress) {
	k.addToIndex(ctx, nodeOperatorIdxKey(consAddr), accAddr.Bytes())
}

func (k Keeper) getNodeOperatorFromIndex(ctx sdk.Context, consAddr sdk.ConsAddress) (sdk.AccAddress, bool) {
	return k.getFromIndex(ctx, nodeOperatorIdxKey(consAddr))
}

func (k Keeper) addProposerToIndex(ctx sdk.Context, height int64, proposer sdk.AccAddress) {
	k.addToIndex(ctx, proposerIdxKey(height), proposer.Bytes())
}

func (k Keeper) getProposerFromIndex(ctx sdk.Context, height int64) (sdk.AccAddress, bool) {
	return k.getFromIndex(ctx, proposerIdxKey(height))
}

func (k Keeper) addToIndex(ctx sdk.Context, key []byte, value []byte) {
	store := ctx.KVStore(k.indexStoreKey)
	store.Set(key, value)
}

func (k Keeper) getFromIndex(ctx sdk.Context, key []byte) (value []byte, found bool) {
	store := ctx.KVStore(k.indexStoreKey)
	if !store.Has(key) {
		return nil, false
	}
	return store.Get(key), true
}
