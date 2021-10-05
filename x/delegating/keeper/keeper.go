package keeper

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/delegating/types"
	"github.com/arterynetwork/artr/x/referral"
)

// Keeper of the delegating store
type Keeper struct {
	mainStoreKey    sdk.StoreKey
	clusterStoreKey sdk.StoreKey
	cdc             codec.BinaryMarshaler
	paramspace      types.ParamSubspace
	accKeeper       types.AccountKeeper
	bankKeeper      types.BankKeeper
	scheduleKeeper  types.ScheduleKeeper
	profileKeeper   types.ProfileKeeper
	refKeeper       types.ReferralKeeper
}

// NewKeeper creates a delegating keeper
func NewKeeper(
	cdc codec.BinaryMarshaler, mainKey sdk.StoreKey, clusterKey sdk.StoreKey, paramspace types.ParamSubspace,
	accountKeeper types.AccountKeeper, scheduleKeeper types.ScheduleKeeper, profileKeeper types.ProfileKeeper,
	bankKeeper types.BankKeeper, refKeeper referral.Keeper,
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
		refKeeper:       refKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

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
		k.cdc.MustUnmarshalBinaryBare(byteItem, &item)
	} else {
		item = types.NewRecord()
	}

	revoking = revoking.Add(uartrs)
	if revoking.GTE(sdk.NewInt(100_000_000000)) {
		if err := ctx.EventManager().EmitTypedEvent(
			&types.EventMassiveRevoke{
				Account: acc.String(),
				Ucoins:  revoking.Uint64(),
			},
		); err != nil { panic(err) }
	}

	nextPayment := ctx.BlockTime().Add(k.scheduleKeeper.OneDay(ctx))
	k.accruePart(ctx, acc, &item, nextPayment)
	if err = k.freeze(ctx, acc, uartrs); err != nil {
		k.Logger(ctx).Error(err.Error())
		return err
	}
	if current.Sub(uartrs).Int64() <= k.bankKeeper.GetParams(ctx).DustDelegation {
		item.NextAccrue = nil
	} else {
		time := ctx.BlockTime().Add(k.scheduleKeeper.OneDay(ctx))
		item.NextAccrue = &time
		k.scheduleKeeper.ScheduleTask(ctx, time, types.AccrueHookName, acc)
	}

	period := k.GetParams(ctx).GetRevokePeriod(k.scheduleKeeper, ctx)
	time := ctx.BlockTime().Add(period)
	item.Requests = append(item.Requests, types.RevokeRequest{
		Time:   time,
		Amount: uartrs,
	})
	store.Set(byteKey, k.cdc.MustMarshalBinaryBare(&item))
	k.scheduleKeeper.ScheduleTask(ctx, time, types.RevokeHookName, byteKey)
	return nil
}

func (k Keeper) Delegate(ctx sdk.Context, acc sdk.AccAddress, uartrs sdk.Int) error {
	if uartrs.LT(sdk.NewInt(k.GetParams(ctx).MinDelegate)) {
		return types.ErrLessThanMinimum
	}

	fee, err := util.PayTxFee(ctx, k.bankKeeper, k.Logger(ctx), acc, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, uartrs)))
	if err != nil {
		return err
	}
	uartrs = uartrs.Sub(fee.AmountOf(util.ConfigMainDenom))

	var (
		store       = ctx.KVStore(k.mainStoreKey)
		byteKey     = []byte(acc)
		nextPayment = ctx.BlockTime().Add(k.scheduleKeeper.OneDay(ctx))

		fees     []referral.ReferralFee
		byteItem []byte
		item     types.Record
	)

	fees, err = k.refKeeper.GetReferralFeesForDelegating(ctx, acc.String())
	if err != nil {
		return err
	}
	k.Logger(ctx).Debug(fmt.Sprintf("Fees: %v", fees))

	totalFee := int64(0)
	outputs := make([]bank.Output, 0, len(fees))
	event := types.EventDelegate{
		Account:          acc.String(),
		CommissionTo:     make([]string, 0, len(fees)),
		CommissionAmount: make([]uint64, 0, len(fees)),
	}
	for _, fee := range fees {
		x := fee.Ratio.MulInt64(uartrs.Int64()).Int64()
		if x == 0 {
			continue
		}
		totalFee += x
		outputs = append(outputs, bank.NewOutput(fee.GetBeneficiary(), sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(x)))))
		event.CommissionTo = append(event.CommissionTo, fee.Beneficiary)
		event.CommissionAmount = append(event.CommissionAmount, uint64(x))
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
		err = proto.Unmarshal(byteItem, &item)
		if err != nil {
			return err
		}
	} else {
		item = types.NewRecord()
	}


	k.accruePart(ctx, acc, &item, nextPayment)
	delegation := uartrs.SubRaw(totalFee)
	if err = k.delegate(ctx, acc, delegation); err != nil {
		return err
	}

	if k.bankKeeper.GetBalance(ctx, acc).AmountOf(util.ConfigDelegatedDenom).Int64() <= k.bankKeeper.GetParams(ctx).DustDelegation {
		item.NextAccrue = nil
	} else {
		time := ctx.BlockTime().Add(k.scheduleKeeper.OneDay(ctx))
		item.NextAccrue = &time
		k.scheduleKeeper.ScheduleTask(ctx, time, types.AccrueHookName, acc)
	}

	event.Ucoins = delegation.Uint64()
	if err := ctx.EventManager().EmitTypedEvent(&event); err != nil { panic(err) }

	bz := k.cdc.MustMarshalBinaryBare(&item)
	store.Set(byteKey, bz)

	return nil
}

func (k Keeper) GetRevoking(ctx sdk.Context, acc sdk.AccAddress) ([]types.RevokeRequest, error) {
	data, err := k.get(ctx, acc)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	return data.Requests, nil
}

func (k Keeper) get(ctx sdk.Context, acc sdk.AccAddress) (*types.Record, error) {
	var (
		store   = ctx.KVStore(k.mainStoreKey)
		byteKey = []byte(acc)

		data types.Record
		err  error
	)
	if !store.Has(byteKey) {
		return nil, nil
	}
	err = proto.Unmarshal(store.Get(byteKey), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (k Keeper) GetAccumulation(ctx sdk.Context, acc sdk.AccAddress) (*types.AccumulationResponse, error) {
	k.Logger(ctx).Debug("GetAccumulation", "acc", acc)
	var (
		store   = ctx.KVStore(k.mainStoreKey)
		byteKey = []byte(acc)

		item types.Record
	)
	if !store.Has(byteKey) {
		return nil, types.ErrNothingDelegated
	}
	k.cdc.MustUnmarshalBinaryBare(store.Get(byteKey), &item)
	if item.NextAccrue == nil {
		return nil, types.ErrNothingDelegated
	}

	periodStart := item.NextAccrue.Add(-k.scheduleKeeper.OneDay(ctx))
	periodEnd := *item.NextAccrue
	dayPart := util.NewFraction(ctx.BlockTime().Sub(periodStart).Nanoseconds(), k.scheduleKeeper.OneDay(ctx).Nanoseconds()).Reduce()

	delegated, _ := k.getDelegated(ctx, acc)
	percent := k.percent(ctx, delegated)
	paymentTotal := percent.MulInt64(delegated.Int64()).Reduce()
	paymentCurrent := paymentTotal.Mul(dayPart)

	result := types.AccumulationResponse{
		Start:         periodStart,
		End:           periodEnd,
		Percent:       percent.MulInt64(100 * 30).Int64(),
		TotalUartrs:   paymentTotal.Int64(),
		CurrentUartrs: paymentCurrent.Int64(),
	}
	k.Logger(ctx).Debug("GetAccumulation", "result", result)
	return &result, nil
}

//----------------------------------------------------------------------------------------------------------------------
// PRIVATE FUNCTIONS



func (k Keeper) getDelegated(ctx sdk.Context, acc sdk.AccAddress) (delegated sdk.Int, undelegating sdk.Int) {
	balance := k.bankKeeper.GetBalance(ctx, acc)
	delegated = balance.AmountOf(util.ConfigDelegatedDenom)
	undelegating = balance.AmountOf(util.ConfigRevokingDenom)
	return
}

func (k Keeper) delegate(ctx sdk.Context, acc sdk.AccAddress, uartrs sdk.Int) error {
	if uartrs.IsZero() {
		return nil
	}
	minusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, uartrs))
	err := k.bankKeeper.SubtractCoins(ctx, acc, minusCoins)

	if err != nil {
		return err
	}

	plusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, uartrs))
	err = k.bankKeeper.AddCoins(ctx, acc, plusCoins)

	if err != nil {
		return err
	}

	supply := k.bankKeeper.GetSupply(ctx)
	supply.Deflate(minusCoins)
	supply.Inflate(plusCoins)
	k.bankKeeper.SetSupply(ctx, supply)

	return nil
}

func (k Keeper) freeze(ctx sdk.Context, acc sdk.AccAddress, uartrds sdk.Int) error {
	if uartrds.IsZero() {
		return nil
	}

	minusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, uartrds))
	err := k.bankKeeper.SubtractCoins(ctx, acc, minusCoins)

	if err != nil {
		return err
	}

	plusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, uartrds))
	err = k.bankKeeper.AddCoins(ctx, acc, plusCoins)

	if err != nil {
		return err
	}

	supply := k.bankKeeper.GetSupply(ctx)
	supply.Deflate(minusCoins)
	supply.Inflate(plusCoins)
	k.bankKeeper.SetSupply(ctx, supply)

	return nil
}

func (k Keeper) undelegate(ctx sdk.Context, acc sdk.AccAddress, uartrs sdk.Int) error {
	if uartrs.IsZero() {
		return nil
	}
	minusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, uartrs))
	err := k.bankKeeper.SubtractCoins(ctx, acc, minusCoins)

	if err != nil {
		return err
	}

	plusCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, uartrs))
	err = k.bankKeeper.AddCoins(ctx, acc, plusCoins)

	if err != nil {
		return err
	}

	supply := k.bankKeeper.GetSupply(ctx)
	supply.Deflate(minusCoins)
	supply.Inflate(plusCoins)
	k.bankKeeper.SetSupply(ctx, supply)

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventUndelegate{
			Account: acc.String(),
			Ucoins:  uartrs.Uint64(),
		},
	); err != nil { panic(err) }
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
	supply := k.bankKeeper.GetSupply(ctx)
	supply.Inflate(emission)
	k.bankKeeper.SetSupply(ctx, supply)

	if err := k.bankKeeper.AddCoins(ctx, acc, emission); err != nil {
		panic(err)
	}
	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventAccrue{
			Account: acc.String(),
			Ucoins:  ucoins.Uint64(),
		},
	); err != nil { panic(err) }
}

func (k Keeper) accruePart(ctx sdk.Context, acc sdk.AccAddress, item *types.Record, nextPayment time.Time) {
	if item.NextAccrue != nil {
		dayPart := k.dayPart(ctx, *item.NextAccrue)
		delegated, _ := k.getDelegated(ctx, acc)
		interest := k.percent(ctx, delegated).Mul(dayPart).Reduce().MulInt64(delegated.Int64()).Int64()
		if interest > 0 {
			k.accrue(ctx, acc, sdk.NewInt(interest))
		}
		k.scheduleKeeper.Delete(ctx, *item.NextAccrue, types.AccrueHookName, acc)
	}
	item.NextAccrue = &nextPayment
}

func (k Keeper) percent(ctx sdk.Context, delegated sdk.Int) util.Fraction {
	var (
		params  = k.GetParams(ctx)
		ladder  = params.Percentage
		percent util.Fraction
	)

	if delegated.Int64() <= k.bankKeeper.GetParams(ctx).DustDelegation {
		return util.FractionZero()
	}

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

func (k Keeper) dayPart(ctx sdk.Context, end time.Time) util.Fraction {
	return util.FractionInt(1).Sub(util.NewFraction(end.Sub(ctx.BlockTime()).Nanoseconds(), k.scheduleKeeper.OneDay(ctx).Nanoseconds()).Reduce())
}
