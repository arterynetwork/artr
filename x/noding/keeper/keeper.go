package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/noding/types"
	"github.com/arterynetwork/artr/x/referral"
)

// Keeper of the noding store
type Keeper struct {
	dataStoreKey     sdk.StoreKey
	indexStoreKey    sdk.StoreKey
	cdc              *codec.Codec
	referralKeeper   types.ReferralKeeper
	scheduleKeeper   types.ScheduleKeeper
	supplyKeeper     types.SupplyKeeper
	paramspace       types.ParamSubspace
	feeCollectorName string
}

// NewKeeper creates a noding keeper
func NewKeeper(
	cdc *codec.Codec,
	dataKey sdk.StoreKey,
	indexKey sdk.StoreKey,
	referralKeeper types.ReferralKeeper,
	scheduleKeeper types.ScheduleKeeper,
	supplyKeeper types.SupplyKeeper,
	paramspace types.ParamSubspace,
	feeCollectorName string,
) Keeper {
	keeper := Keeper{
		dataStoreKey:     dataKey,
		indexStoreKey:    indexKey,
		cdc:              cdc,
		referralKeeper:   referralKeeper,
		scheduleKeeper:   scheduleKeeper,
		supplyKeeper:     supplyKeeper,
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

func (k Keeper) IsQualified(ctx sdk.Context, accAddr sdk.AccAddress) (result bool, delegation sdk.Int, reason string, err error) {
	// Check if it's staff
	if k.has(ctx, accAddr) {
		var d types.D
		d, err = k.Get(ctx, accAddr)
		if err != nil {
			return
		}
		if d.Staff {
			result = true
		}
	}

	// Check minimal status
	status, err := k.referralKeeper.GetStatus(ctx, accAddr)
	if err != nil {
		return
	}
	if !result && status < referral.Status(k.GetParams(ctx).MinStatus) {
		reason = types.AttributeValueNotEnoughStatus
		return
	}

	// 10k ARTR delegated
	delegation, err = k.referralKeeper.GetDelegatedInNetwork(ctx, accAddr, 10)
	if err != nil {
		return
	}
	if !result && delegation.Int64() < 10_000_000000 {
		reason = types.AttributeValueNotEnoughDelegation
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
		var data types.D
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

	power := k.power(ctx, delegation.Int64())
	if k.has(ctx, accAddr) {
		err = k.update(ctx, accAddr, func(d *types.D) (save bool) {
			d.Status = true
			d.PubKey = bech32FromCryptoPubKey(key)
			d.Power = power
			return true
		})
	} else {
		err = k.set(ctx, accAddr, types.NewD(power, bech32FromCryptoPubKey(key)))
	}
	if err != nil {
		return err
	}

	k.addToIndex(ctx, nodeOperatorIdxKey(consAddressFromCryptoBubKey(key)), accAddr.Bytes())
	return nil
}

func (k Keeper) SwitchOff(ctx sdk.Context, accAddr sdk.AccAddress) error {
	err := k.update(ctx, accAddr, func(d *types.D) (save bool) {
		if d.Power == 0 && !d.Status {
			return false
		}

		d.Power = 0
		d.Status = false
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

	var reason string
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

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeValidatorBanished,
		sdk.NewAttribute(types.AttributeKeyAccountAddress, acc.String()),
		sdk.NewAttribute(types.AttributeKeyReason, reason),
	))
	return nil
}

func (k Keeper) OnStakeChanged(ctx sdk.Context, acc sdk.AccAddress) error {
	is, err := k.IsValidator(ctx, acc)
	if err != nil {
		return err
	}
	if !is {
		return nil
	}

	record, err := k.Get(ctx, acc)
	if err != nil {
		return err
	}

	is, delegation, reason, err := k.IsQualified(ctx, acc)
	if err != nil {
		return err
	}
	if !is {
		k.Logger(ctx).Info("not qualified anymore, banishing from validators", "account", acc)
		err = k.SwitchOff(ctx, acc)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeValidatorBanished,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, acc.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		))
		return nil
	}

	newPower := k.power(ctx, delegation.Int64())
	dPower := newPower - record.Power
	if dPower == 0 {
		return nil
	}

	record.Power = newPower
	return k.set(ctx, acc, record)
}

func (k Keeper) GatherValidatorUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	var (
		store = ctx.KVStore(k.dataStoreKey)

		result []abci.ValidatorUpdate
		active []types.KeyedD
	)

	it := store.Iterator(nil, nil)
	for ; it.Valid(); it.Next() {
		var (
			addr sdk.AccAddress
			data types.D
		)
		addr = sdk.AccAddress(it.Key())
		k.cdc.MustUnmarshalBinaryLengthPrefixed(it.Value(), &data)
		if data.IsActive() {
			active = append(active, types.NewKeyedD(addr, data))
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
				if err := k.update(ctx, addr, func(d *types.D) (save bool) {
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
		sort.Slice(active, func(i, j int) bool {
			xi := active[i]
			xj := active[j]

			if xi.Strokes < xj.Strokes {
				return true
			}
			if xi.Strokes > xj.Strokes {
				return false
			}

			if xi.Power > xj.Power {
				return true
			}
			if xi.Power < xj.Power {
				return false
			}

			return xi.OkBlocksInRow > xj.OkBlocksInRow
		})
	} else {
		n1 = len(active)
		n2 = 0
		// No one will be left out, so all're equal. No need to sort items.
	}

	for i = 0; i < n1; i++ {
		data := active[i]
		if len(data.PubKey) == 0 {
			panic("validator cannot be active without PubKey")
		}

		updated := data.LastPower != data.Power
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
				Power:  data.Power,
			})
		}
		if data.LotteryNo != 0 {
			if err := k.lotteryExclude(ctx, &data.D); err != nil {
				// Should never happen, we've just checked they has been participating
				panic(err)
			}
			updated = true
		}
		if updated {
			if err := k.update(ctx, data.Account, func(d *types.D) (save bool) {
				d.LastPower = d.Power
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
			if data.PubKey != data.LastPubKey {
				if len(data.LastPubKey) != 0 {
					result = append(result, abci.ValidatorUpdate{
						PubKey: abciPubKeyFromBech32(data.LastPubKey),
						Power:  0,
					})
				}
			} else if data.LastPower == data.Power {
				continue
			}

			result = append(result, abci.ValidatorUpdate{
				PubKey: abciPubKeyFromBech32(data.PubKey),
				Power:  data.Power,
			})
			if err := k.update(ctx, data.Account, func(d *types.D) (save bool) {
				d.LastPower = d.Power
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
				if err := k.lotteryAddNew(ctx, data.Account, &data.D); err != nil {
					// Should never happen, we've just checked that the account hasn't been participating yet.
					panic(err)
				}
				updated = true
			}
			if updated {
				if err := k.update(ctx, data.Account, func(d *types.D) (save bool) {
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
			if bytes.Equal(x.PubKey.Data, y.PubKey.Data) {
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
		var value types.D
		if err := k.cdc.UnmarshalBinaryLengthPrefixed(it.Value(), &value); err != nil {
			return nil, err
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
		var value types.D
		if err := k.cdc.UnmarshalBinaryLengthPrefixed(it.Value(), &value); err != nil {
			return nil, err
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
		var value types.D
		if err := k.cdc.UnmarshalBinaryLengthPrefixed(it.Value(), &value); err != nil {
			return nil, err
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
		pubkey := cryptoPubKeyFromBech32(v.Pubkey)

		if err := k.set(ctx, v.Account, v.ToD()); err != nil {
			return errors.Wrap(err, "cannot set data")
		}
		for _, h := range v.ProposedBlocks {
			k.addProposerToIndex(ctx, int64(h), v.Account)
		}
		if err := k.SwitchOn(ctx, v.Account, pubkey); err != nil {
			return errors.Wrap(err, "cannot switch on")
		}
	}
	return nil
}

func (k Keeper) SetNonActiveValidators(ctx sdk.Context, validators []types.Validator) error {
	for _, v := range validators {
		if err := k.set(ctx, v.Account, v.ToD()); err != nil {
			return err
		}
		for _, h := range v.ProposedBlocks {
			k.addProposerToIndex(ctx, int64(h), v.Account)
		}
	}
	return nil
}

// MarkStroke - to be called every time the validator misses a block.
func (k Keeper) MarkStroke(ctx sdk.Context, acc sdk.AccAddress) error {
	p := k.GetParams(ctx)

	return k.update(ctx, acc, func(d *types.D) (save bool) {
		if d.Jailed {
			return false
		}

		d.Strokes++
		d.OkBlocksInRow = 0
		d.MissedBlocksInRow++
		if d.LotteryNo != 0 {
			if err := k.lotteryDownshift(ctx, acc, d); err != nil {
				// Should never happen
				panic(err)
			}
		}
		if d.MissedBlocksInRow >= int64(p.JailAfter) {
			d.Power = 0
			d.Jailed = true
			d.UnjailAt = ctx.BlockHeight() + p.UnjailAfter
			d.JailCount++
			d.MissedBlocksInRow = 0
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeValidatorJailed,
				sdk.NewAttribute(types.AttributeKeyAccountAddress, acc.String()),
			))
		}
		return true
	})
}

// MarkTick - to be called every time the validator signs a block successfully.
func (k Keeper) MarkTick(ctx sdk.Context, acc sdk.AccAddress) error {
	return k.update(ctx, acc, func(d *types.D) (save bool) {
		d.MissedBlocksInRow = 0
		d.OkBlocksInRow++
		return true
	})
}

func (k Keeper) MarkByzantine(ctx sdk.Context, acc sdk.AccAddress, evidence abci.Evidence) error {
	return k.update(ctx, acc, func(d *types.D) (save bool) {
		var eventType string
		d.Infractions = append(d.Infractions, evidence)
		if len(d.Infractions) > 1 {
			d.BannedForLife = true
			d.Status = false
			d.Power = 0
			eventType = types.EventTypeValidatorBanned
		} else {
			eventType = types.EventTypeValidatorWarning
		}
		ctx.EventManager().EmitEvent(sdk.NewEvent(eventType,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, acc.String()),
			sdk.NewAttribute(types.AttributeKeyEvidences, fmt.Sprintf("%+v", d.Infractions)),
		))
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
		data.Power = k.power(ctx, delegation.Int64())
	} else {
		k.Logger(ctx).Info("banishing from validators", "acc", acc, "reason", reason)
		data.Status = false
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeValidatorBanished,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, acc.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		))
	}
	err = k.set(ctx, acc, data)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) PayProposerReward(ctx sdk.Context, acc sdk.AccAddress) (err error) {
	k.addProposerToIndex(ctx, ctx.BlockHeight()-1, acc)
	if err := k.update(ctx, acc, func(d *types.D) (save bool) {
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

	all := k.supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName).GetCoins()
	amount := all
	if amount.IsZero() {
		return nil
	}
	if err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, acc, amount); err != nil {
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
		err = k.update(ctx, acc, func(d *types.D) (save bool) {
			if d.Staff {
				return false
			}

			d.Staff = true
			return true
		})
	} else {
		err = k.set(ctx, acc, types.D{Staff: true})
	}
	return err
}

func (k Keeper) RemoveFromStaff(ctx sdk.Context, acc sdk.AccAddress) (err error) {
	if !k.has(ctx, acc) {
		return nil
	}

	var isActive bool
	err = k.update(ctx, acc, func(d *types.D) (save bool) {
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
		var item types.D
		k.cdc.MustUnmarshalBinaryLengthPrefixed(it.Value(), &item)
		if item.Strokes == 0 && item.JailCount == 0 {
			continue
		}
		item.Strokes = 0
		item.JailCount = 0
		store.Set(it.Key(), k.cdc.MustMarshalBinaryLengthPrefixed(item))
	}
}

func (k Keeper) GetValidatorState(ctx sdk.Context, acc sdk.AccAddress) types.ValidatorState {
	data, err := k.Get(ctx, acc)
	if err != nil {
		return types.ValidatorStateOff
	}
	if data.BannedForLife {
		return types.ValidatorStateBan
	}
	if data.Jailed {
		return types.ValidatorStateJail
	}
	if !data.Status {
		return types.ValidatorStateOff
	}
	if data.LastPower == 0 {
		return types.ValidatorStateSpare
	}
	if data.LotteryNo == 0 {
		return types.ValidatorStateTop
	}
	return types.ValidatorStateLucky
}

//----------------------------------------------------------------------------------

func (k Keeper) has(ctx sdk.Context, acc sdk.AccAddress) bool {
	return ctx.KVStore(k.dataStoreKey).Has(acc)
}

func (k Keeper) Get(ctx sdk.Context, acc sdk.AccAddress) (types.D, error) {
	store := ctx.KVStore(k.dataStoreKey)
	key := []byte(acc)
	if !store.Has(key) {
		return types.D{}, types.ErrNotFound
	}
	var item types.D
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(key), &item)
	return item, err
}

func (k Keeper) set(ctx sdk.Context, acc sdk.AccAddress, value types.D) error {
	store := ctx.KVStore(k.dataStoreKey)
	keyBytes := []byte(acc)
	valueBytes, err := k.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}

	store.Set(keyBytes, valueBytes)
	return nil
}

func (k Keeper) update(ctx sdk.Context, acc sdk.AccAddress, callback func(d *types.D) (save bool)) error {
	var (
		store      = ctx.KVStore(k.dataStoreKey)
		keyBytes   = []byte(acc)
		value      types.D
		valueBytes []byte
		err        error
	)
	if !store.Has(keyBytes) {
		return types.ErrNotFound
	}
	err = k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(keyBytes), &value)
	if err != nil {
		return err
	}

	if !callback(&value) {
		return nil
	}
	valueBytes, err = k.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}

	store.Set(keyBytes, valueBytes)
	return nil
}

func (k Keeper) power(_ sdk.Context, delegated int64) int64 {
	if delegated >= 100_000_000000 {
		return 15
	} else {
		return 10
	}
}

func abciPubKeyFromBech32(bech32 string) abci.PubKey {
	return tmtypes.TM2PB.PubKey(cryptoPubKeyFromBech32(bech32))
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
	if height > k.scheduleKeeper.GetParams(ctx).InitialHeight {
		k.addToIndex(ctx, proposerIdxKey(height), proposer.Bytes())
	}
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
