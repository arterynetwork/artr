package keeper

import (
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/referral"
	"encoding/binary"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"sort"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/delegating/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the delegating store
type Keeper struct {
	mainStoreKey    sdk.StoreKey
	clusterStoreKey sdk.StoreKey
	cdc             *codec.Codec
	paramspace      types.ParamSubspace
	accKeeper       types.AccountKeeper
	bankKeeper      types.BankKeeper
	supplyKeeper    types.SupplyKeeper
	scheduleKeeper  types.ScheduleKeeper
	profileKeeper   types.ProfileKeeper
	refKeeper       types.ReferralKeeper
}

// NewKeeper creates a delegating keeper
func NewKeeper(
	cdc *codec.Codec, mainKey sdk.StoreKey, clusterKey sdk.StoreKey, paramspace types.ParamSubspace,
	accountKeeper types.AccountKeeper, scheduleKeeper types.ScheduleKeeper, profileKeeper types.ProfileKeeper,
	bankKeeper types.BankKeeper, supplyKeeper types.SupplyKeeper, refKeeper referral.Keeper,
) Keeper {
	keeper := Keeper{
		mainStoreKey:    mainKey,
		clusterStoreKey: clusterKey,
		cdc:             cdc,
		paramspace:      paramspace.WithKeyTable(types.ParamKeyTable()),
		accKeeper:       accountKeeper,
		scheduleKeeper:  scheduleKeeper,
		profileKeeper:   profileKeeper,
		bankKeeper:      bankKeeper,
		supplyKeeper:    supplyKeeper,
		refKeeper:       refKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

const never = -1 // in terms of a block height
const oneDay = util.BlocksOneDay
const twoWeeks = 14 * oneDay

func (k Keeper) Revoke(ctx sdk.Context, acc sdk.AccAddress, uartrs sdk.Int) error {
	if uartrs.IsZero() {
		return nil
	}
	var (
		store             = ctx.KVStore(k.mainStoreKey)
		byteKey           = []byte(acc)
		current, revoking = k.getDelegated(ctx, acc)

		byteItem []byte
		item     types.Record
		err      error
	)

	if uartrs.GT(current) {
		err = sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "cannot revoke from delegation more than delegated")
		k.Logger(ctx).Error(err.Error())
		return err
	}

	if store.Has(byteKey) {
		byteItem = store.Get(byteKey)
		err = k.cdc.UnmarshalBinaryLengthPrefixed(byteItem, &item)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			return err
		}
	} else {
		item = types.NewRecord()
	}

	revoking = revoking.Add(uartrs)
	if revoking.GTE(sdk.NewInt(100_000_000000)) {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMassiveRevoke,
				sdk.NewAttribute(types.AttributeKeyAccount, acc.String()),
				sdk.NewAttribute(types.AttributeKeyUcoins, revoking.String()),
			),
		)
	}

	nextPayment := ctx.BlockHeight() + oneDay
	if err = k.accruePart(ctx, acc, &item, nextPayment); err != nil {
		return err
	}
	if err = k.freeze(ctx, acc, uartrs); err != nil {
		k.Logger(ctx).Error(err.Error())
		return err
	}
	if uartrs.Equal(current) {
		item.Cluster = never
	} else {
		err = k.addToCluster(ctx, item.Cluster, acc)
		if err != nil {
			return err
		}
	}

	height := ctx.BlockHeight() + twoWeeks
	item.Requests = append(item.Requests, types.RevokeRequest{
		HeightToImplementAt: height,
		MicroCoins:          uartrs,
	})
	byteItem, err = k.cdc.MarshalBinaryLengthPrefixed(item)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return err
	}
	store.Set(byteKey, byteItem)
	err = k.scheduleKeeper.ScheduleTask(ctx, uint64(height), types.RevokeHookName, &byteKey)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return err
	}
	return nil
}

func (k Keeper) MustPerformRevoking(ctx sdk.Context, payload []byte) {
	if err := k.performRevoking(ctx, payload); err != nil {
		panic(err)
	}
}

func (k Keeper) Delegate(ctx sdk.Context, acc sdk.AccAddress, uartrs sdk.Int) error {
	if uartrs.IsZero() {
		return nil
	}
	var (
		store       = ctx.KVStore(k.mainStoreKey)
		byteKey     = []byte(acc)
		nextPayment = ctx.BlockHeight() + oneDay

		fees     []referral.ReferralFee
		byteItem []byte
		item     types.Record
		err      error
	)

	fees, err = k.refKeeper.GetReferralFeesForDelegating(ctx, acc)
	if err != nil {
		return err
	}
	k.Logger(ctx).Debug(fmt.Sprintf("Fees: %v", fees))

	totalFee := int64(0)
	outputs := make([]bank.Output, 0, len(fees))
	eAttrs := make([]sdk.Attribute, 0, 2*len(fees)+2)
	eAttrs = append(eAttrs,
		sdk.NewAttribute(types.AttributeKeyAccount, acc.String()),
		sdk.NewAttribute(types.AttributeKeyUcoins, "" /* will set later */),
	)
	for _, fee := range fees {
		x := fee.Ratio.MulInt64(uartrs.Int64()).Int64()
		if x == 0 {
			continue
		}
		totalFee += x
		outputs = append(outputs, bank.NewOutput(fee.Beneficiary, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(x)))))
		eAttrs = append(eAttrs,
			sdk.NewAttribute(types.AttributeKeyCommissionTo, fee.Beneficiary.String()),
			sdk.NewAttribute(types.AttributeKeyCommissionAmount, fmt.Sprintf("%d", x)),
		)
	}
	if totalFee != 0 {
		inputs := []bank.Input{bank.NewInput(acc, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(totalFee))))}

		err = k.bankKeeper.InputOutputCoins(ctx, inputs, outputs)
		if err != nil {
			return err
		}
	}

	if store.Has(byteKey) {
		byteItem = store.Get(byteKey)
		err = k.cdc.UnmarshalBinaryLengthPrefixed(byteItem, &item)
		if err != nil {
			return err
		}
	} else {
		item = types.NewRecord()
	}

	if err = k.accruePart(ctx, acc, &item, nextPayment); err != nil {
		return err
	}
	delegation := uartrs.SubRaw(totalFee)
	if err = k.delegate(ctx, acc, delegation); err != nil {
		return err
	}

	eAttrs[1].Value = delegation.String()

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeDelegate, eAttrs...))

	store.Set(byteKey, k.cdc.MustMarshalBinaryLengthPrefixed(item))
	return k.addToCluster(ctx, item.Cluster, acc)
}

func (k Keeper) Accrue(ctx sdk.Context) error {
	height := ctx.BlockHeight()
	if height <= k.scheduleKeeper.GetParams(ctx).InitialHeight {
		k.Logger(ctx).Debug("Accrue: fast-forward, doing nothing")
		return nil
	}
	var (
		store   = ctx.KVStore(k.clusterStoreKey)
		byteKey = getCluster(height)

		targets []sdk.AccAddress
		err     error
	)
	if !store.Has(byteKey) {
		k.Logger(ctx).Debug("Accrue: nothing scheduled", "cluster", fmt.Sprintf("%v", byteKey))
		return nil
	}
	err = k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(byteKey), &targets)
	if err != nil || targets == nil {
		return err
	}

	store = ctx.KVStore(k.mainStoreKey)
	for _, acc := range targets {
		delegated, _ := k.getDelegated(ctx, acc)
		percent := k.percent(ctx, delegated)
		k.accrue(ctx, acc, sdk.NewInt(percent.MulInt64(delegated.Int64()).Int64()))
	}
	k.Logger(ctx).Debug("Accrue", "count", len(targets))
	return nil
}

func (k Keeper) GetRevoking(ctx sdk.Context, acc sdk.AccAddress) ([]types.RevokeRequest, error) {
	var (
		store   = ctx.KVStore(k.mainStoreKey)
		byteKey = []byte(acc)

		data types.Record
		err  error
	)
	if !store.Has(byteKey) {
		return nil, nil
	}
	err = k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(byteKey), &data)
	if err != nil {
		return nil, err
	}

	return data.Requests, nil
}

func (k Keeper) GetAccumulation(ctx sdk.Context, acc sdk.AccAddress) (types.QueryResAccumulation, error) {
	k.Logger(ctx).Debug("GetAccumulation", "acc", acc)
	var (
		store   = ctx.KVStore(k.mainStoreKey)
		byteKey = []byte(acc)

		item types.Record
		err  error
	)
	if !store.Has(byteKey) {
		return types.QueryResAccumulation{}, sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, "nothing's delegated (A)")
	}
	err = k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(byteKey), &item)
	if err != nil {
		return types.QueryResAccumulation{}, err
	}
	if item.Cluster == never {
		return types.QueryResAccumulation{}, sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, "nothing's delegated (B)")
	}

	periodStart := ctx.BlockHeight() - (ctx.BlockHeight()-item.Cluster)%oneDay
	periodEnd := periodStart + oneDay
	dayPart := util.NewFraction(ctx.BlockHeight()-periodStart, oneDay)

	delegated, _ := k.getDelegated(ctx, acc)
	percent := k.percent(ctx, delegated)
	paymentTotal := percent.MulInt64(delegated.Int64()).Reduce()
	paymentCurrent := paymentTotal.Mul(dayPart)

	result := types.QueryResAccumulation{
		StartHeight:   periodStart,
		EndHeight:     periodEnd,
		Percent:       int(percent.MulInt64(100 * 30).Int64()),
		TotalUartrs:   paymentTotal.Int64(),
		CurrentUartrs: paymentCurrent.Int64(),
	}
	k.Logger(ctx).Debug("GetAccumulation", "result", result)
	return result, nil
}

//----------------------------------------------------------------------------------------------------------------------
// PRIVATE FUNCTIONS

func (k Keeper) performRevoking(ctx sdk.Context, payload []byte) error {
	var (
		acc       = sdk.AccAddress(payload)
		mainStore = ctx.KVStore(k.mainStoreKey)
		byteKey   = []byte(acc)

		byteItem []byte
		item     types.Record
		err      error
	)

	if !mainStore.Has(byteKey) {
		return nil
	}
	byteItem = mainStore.Get(byteKey)
	if err = k.cdc.UnmarshalBinaryLengthPrefixed(byteItem, &item); err != nil {
		return err
	}
	sort.Slice(item.Requests, func(i, j int) bool {
		return item.Requests[i].HeightToImplementAt < item.Requests[j].HeightToImplementAt
	})

	n := 0
	for _, req := range item.Requests {
		if req.HeightToImplementAt > ctx.BlockHeight() {
			break
		}

		if err = k.undelegate(ctx, acc, req.MicroCoins); err != nil {
			return err
		}
		n += 1
	}
	if n == 0 {
		return nil
	}
	item.Requests = item.Requests[n:]
	if item.IsEmpty() {
		mainStore.Delete(byteKey)
	} else {
		mainStore.Set(byteKey, k.cdc.MustMarshalBinaryLengthPrefixed(item))
	}
	return nil
}

func (k Keeper) getDelegated(ctx sdk.Context, acc sdk.AccAddress) (delegated sdk.Int, undelegating sdk.Int) {
	coins := k.accKeeper.GetAccount(ctx, acc).GetCoins()

	delegated = coins.AmountOf(util.ConfigDelegatedDenom)
	undelegating = coins.AmountOf(util.ConfigRevokingDenom)
	return
}

func (k Keeper) delegate(ctx sdk.Context, acc sdk.AccAddress, uartrs sdk.Int) error {
	if uartrs.IsZero() {
		return nil
	}
	minusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, uartrs))
	_, err := k.bankKeeper.SubtractCoins(ctx, acc, minusCoins)

	if err != nil {
		return err
	}

	plusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, uartrs))
	_, err = k.bankKeeper.AddCoins(ctx, acc, plusCoins)

	if err != nil {
		return err
	}

	supply := k.supplyKeeper.GetSupply(ctx)
	supply = supply.Deflate(minusCoins).Inflate(plusCoins)
	k.supplyKeeper.SetSupply(ctx, supply)

	return nil
}

func (k Keeper) freeze(ctx sdk.Context, acc sdk.AccAddress, uartrds sdk.Int) error {
	if uartrds.IsZero() {
		return nil
	}

	minusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, uartrds))
	_, err := k.bankKeeper.SubtractCoins(ctx, acc, minusCoins)

	if err != nil {
		return err
	}

	plusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, uartrds))
	_, err = k.bankKeeper.AddCoins(ctx, acc, plusCoins)

	if err != nil {
		return err
	}

	supply := k.supplyKeeper.GetSupply(ctx)
	supply = supply.Deflate(minusCoins).Inflate(plusCoins)
	k.supplyKeeper.SetSupply(ctx, supply)

	return nil
}

func (k Keeper) undelegate(ctx sdk.Context, acc sdk.AccAddress, uartrs sdk.Int) error {
	if uartrs.IsZero() {
		return nil
	}
	minusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, uartrs))
	_, err := k.bankKeeper.SubtractCoins(ctx, acc, minusCoins)

	if err != nil {
		return err
	}

	plusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, uartrs))
	_, err = k.bankKeeper.AddCoins(ctx, acc, plusCoins)

	if err != nil {
		return err
	}

	supply := k.supplyKeeper.GetSupply(ctx)
	supply = supply.Deflate(minusCoins).Inflate(plusCoins)
	k.supplyKeeper.SetSupply(ctx, supply)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUndelegate,
		sdk.NewAttribute(types.AttributeKeyAccount, acc.String()),
		sdk.NewAttribute(types.AttributeKeyUcoins, uartrs.String()),
	))
	return nil
}

func (k Keeper) accrue(ctx sdk.Context, acc sdk.AccAddress, ucoins sdk.Int) {
	if ucoins.IsZero() {
		return
	}

	profile := k.profileKeeper.GetProfile(ctx, acc)
	if profile == nil {
		k.Logger(ctx).Error("profile not found, not accruing", "acc", acc)
		return
	}

	emission := sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, ucoins))
	supply := k.supplyKeeper.GetSupply(ctx)

	k.supplyKeeper.SetSupply(ctx, supply.Inflate(emission))

	_, err := k.bankKeeper.AddCoins(ctx, acc, emission)
	if err != nil {
		panic(err)
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeAccrue,
		sdk.NewAttribute(types.AttributeKeyAccount, acc.String()),
		sdk.NewAttribute(types.AttributeKeyUcoins, ucoins.String()),
	))
}

func (k Keeper) accruePart(ctx sdk.Context, acc sdk.AccAddress, item *types.Record, nextPayment int64) error {
	if item.Cluster != never {
		dayPart := util.NewFraction((nextPayment-item.Cluster)%oneDay, oneDay)
		delegated, _ := k.getDelegated(ctx, acc)
		interest := k.percent(ctx, delegated).Mul(dayPart).Reduce().MulInt64(delegated.Int64()).Int64()
		if interest > 0 {
			k.accrue(ctx, acc, sdk.NewInt(interest))
		}
		if err := k.dropFromCluster(ctx, item.Cluster, acc); err != nil {
			return err
		}
	}
	item.Cluster = nextPayment % oneDay
	return nil
}

func (k Keeper) percent(ctx sdk.Context, delegated sdk.Int) util.Fraction {
	var (
		params  = k.GetParams(ctx)
		ladder  = params.Percentage
		percent util.Fraction
	)

	if delegated.LT(sdk.NewInt(1_000_000000)) {
		percent = util.Percent(int64(ladder.Minimal))
	} else if delegated.LT(sdk.NewInt(10_000_000000)) {
		percent = util.Percent(int64(ladder.ThousandPlus))
	} else if delegated.LT(sdk.NewInt(100_000_000000)) {
		percent = util.Percent(int64(ladder.TenKPlus))
	} else {
		percent = util.Percent(int64(ladder.HundredKPlus))
	}
	percent = percent.Div(util.NewFraction(30, 1)) // to days from months
	return percent.Reduce()
}

func getCluster(height int64) []byte {
	if height < 0 {
		return nil
	}
	res := make([]byte, 2)
	binary.BigEndian.PutUint16(res, uint16(height%oneDay))
	return res
}

func (k Keeper) addToCluster(ctx sdk.Context, cluster int64, acc sdk.AccAddress) error {
	var (
		store = ctx.KVStore(k.clusterStoreKey)
		key   = make([]byte, 2)

		err  error
		buf  []byte
		data []sdk.AccAddress
	)
	binary.BigEndian.PutUint16(key, uint16(cluster))
	if store.Has(key) {
		buf = store.Get(key)
		err = k.cdc.UnmarshalBinaryLengthPrefixed(buf, &data)
		if err != nil {
			return err
		}
	} else {
		data = nil
	}
	data = append(data, acc)
	buf, err = k.cdc.MarshalBinaryLengthPrefixed(data)
	if err != nil {
		return err
	}
	store.Set(key, buf)
	k.Logger(ctx).Debug("account added to cluster", "acc", acc, "cluster", cluster)
	return nil
}

func (k Keeper) dropFromCluster(ctx sdk.Context, cluster int64, acc sdk.AccAddress) error {
	var (
		store = ctx.KVStore(k.clusterStoreKey)
		key   = make([]byte, 2)

		err  error
		buf  []byte
		data []sdk.AccAddress
	)
	binary.BigEndian.PutUint16(key, uint16(cluster))

	buf = store.Get(key)
	err = k.cdc.UnmarshalBinaryLengthPrefixed(buf, &data)
	if err != nil {
		return err
	}

	for i, x := range data {
		if x.Equals(acc) {
			data[i] = data[len(data)-1]
			data = data[:len(data)-1]
			break
		}
	}
	if data == nil {
		store.Delete(key)
	} else {
		buf, err = k.cdc.MarshalBinaryLengthPrefixed(data)
		if err != nil {
			return err
		}
		store.Set(key, buf)
	}
	k.Logger(ctx).Debug("account dropped from cluster", "acc", acc, "cluster", cluster)
	return nil
}
