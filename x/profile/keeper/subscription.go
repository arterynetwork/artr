package keeper

import (
	"math/big"
	"time"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/earning"
	"github.com/arterynetwork/artr/x/profile/types"
)

func (k Keeper) PayTariff(ctx sdk.Context, addr sdk.AccAddress, storageGb uint32) error {
	p := k.GetParams(ctx)
	profile := k.GetProfile(ctx, addr)

	var storageB uint64
	if storageGb == 0 {
		storageB = profile.StorageLimit
	} else {
		storageB = uint64(storageGb) * util.GBSize
		if storageGb < p.BaseStorageGb {
			return errors.Errorf("storage amount is below minimum (%d < %d)", storageGb, p.BaseStorageGb)
		}
		if storageB < profile.StorageCurrent {
			return errors.Errorf("storage amount is below current consumption (%dG < %d)", storageGb, profile.StorageCurrent)
		}
	}
	// NOTE: We shouldn't use `storageGb` below this point in case it's zero.

	tariffTotal := sdk.NewIntFromBigInt(p.TokenRate.MulInt64(int64(p.SubscriptionPrice)).BigInt())
	// NOTE: `tariffTotal` cannot be just assigned to `total` here, 'cause Int is a struct over a pointer.
	total := sdk.NewIntFromBigInt(new(big.Int).Set(tariffTotal.BigInt()))

	txFee := util.CalculateFee(tariffTotal)
	tariffTotal = tariffTotal.Sub(txFee)

	if refInfo, err := k.referralKeeper.Get(ctx, addr.String()); err != nil {
		return errors.Wrap(err, "cannot obtain referral data")
	} else if refInfo.Banished {
		k.Logger(ctx).Info("account is banished, turning it back", "address", addr.String())
		if err := k.referralKeeper.ComeBack(ctx, addr.String()); err != nil {
			return errors.Wrap(err, "cannot return a banished account")
		}
	}

	refFees, err := k.referralKeeper.GetReferralFeesForSubscription(ctx, addr.String())
	if err != nil {
		return errors.Wrap(err, "cannot calculate referral fees")
	}
	refTotal := int64(0)
	outputs := make([]bank.Output, len(refFees), len(refFees)+3)
	for i, fee := range refFees {
		x := fee.Ratio.MulInt64(tariffTotal.Int64()).Int64()
		refTotal += x
		outputs[i] = bank.NewOutput(fee.GetBeneficiary(), util.Uartrs(x))
	}
	tariffTotal = tariffTotal.SubRaw(refTotal)

	storageFeeFrac := p.TokenRate.
		MulInt64((int64(storageB) - int64(p.BaseStorageGb)*util.GBSize) * int64(p.StorageGbPrice)).
		DivInt64(util.GBSize).Reduce()
	if !profile.IsActive(ctx) {
		au := ctx.BlockTime().Add(k.scheduleKeeper.OneMonth(ctx))
		profile.ActiveUntil = &au
		k.scheduleRenew(ctx, addr, au)
		k.resetLimits(p, profile)
		if err := ctx.EventManager().EmitTypedEvent(
			&types.EventActivityChanged{
				Address:   addr.String(),
				ActiveNow: true,
			},
		); err != nil { panic(err) }
	} else {
		if storageB != profile.StorageLimit {
			time := k.monthPart(ctx, *profile.ActiveUntil)
			storageFeeFrac = storageFeeFrac.Add(
				p.TokenRate.
					MulInt64(int64(storageB) - int64(profile.StorageLimit)).
					DivInt64(util.GBSize).Reduce().
					MulInt64(int64(p.StorageGbPrice)).
					Mul(time).Reduce(),
			)
			if storageFeeFrac.IsNegative() {
				storageFeeFrac = util.FractionZero()
			}
			profile.StorageLimit = storageB
		}

		*profile.ActiveUntil = profile.ActiveUntil.Add(k.scheduleKeeper.OneMonth(ctx))
	}

	vpnFee := tariffTotal.QuoRaw(3)

	storageFee := sdk.NewIntFromBigInt(storageFeeFrac.BigInt())
	total = total.Add(storageFee)
	storageFee = storageFee.Add(tariffTotal.Sub(vpnFee))

	if !txFee.IsZero() {
		outputs = append(outputs, bank.NewOutput(k.accountKeeper.GetModuleAddress(auth.FeeCollectorName), sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, txFee))))
	}
	if !vpnFee.IsZero() {
		outputs = append(outputs, bank.NewOutput(k.accountKeeper.GetModuleAddress(earning.VpnCollectorName), sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, vpnFee))))
	}
	if !storageFee.IsZero() {
		outputs = append(outputs, bank.NewOutput(k.accountKeeper.GetModuleAddress(earning.StorageCollectorName), sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, storageFee))))
	}

	input := bank.NewInput(addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, total)))
	if err := k.bankKeeper.InputOutputCoins(ctx,
		[]bank.Input{input},
		outputs,
	); err != nil {
		return errors.Wrap(err, "cannot pay up fees")
	}
	if err := k.SetProfile(ctx, addr, *profile); err != nil {
		return errors.Wrap(err, "cannot save profile")
	}

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventPayTariff{
			Address:  addr.String(),
			ExpireAt: *profile.ActiveUntil,
			Total:    input.Coins.AmountOf(util.ConfigMainDenom).Uint64(),
		},
	); err != nil { return err }
	return nil
}

func (k Keeper) BuyStorage(ctx sdk.Context, addr sdk.AccAddress, extraGb uint32) error {
	profile := k.GetProfile(ctx, addr)
	if !profile.IsActive(ctx) {
		return errors.New("account is not active")
	}

	p := k.GetParams(ctx)
	time := k.monthPart(ctx, *profile.ActiveUntil)
	storageFee := p.TokenRate.MulInt64(int64(extraGb) * int64(p.StorageGbPrice)).Mul(time).Int64()

	if storageFee <= 0 {
		panic("free storage")
	}
	total := util.Uartrs(storageFee)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, earning.StorageCollectorName, total); err != nil {
		return errors.Wrap(err, "cannot pay up fee")
	}
	profile.StorageLimit += uint64(extraGb) * util.GBSize
	if err := k.SetProfile(ctx, addr, *profile); err != nil {
		return errors.Wrap(err, "cannot write profile")
	}
	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventBuyStorage{
			Address:  addr.String(),
			NewLimit: profile.StorageLimit,
			Used:     profile.StorageCurrent,
			Total:    total.AmountOf(util.ConfigMainDenom).Uint64(),
		},
	); err != nil { panic(err) }
	return nil
}

func (k Keeper) BuyVpn(ctx sdk.Context, addr sdk.AccAddress, vpnGb uint32) error {
	profile := k.GetProfile(ctx, addr)
	if !profile.IsActive(ctx) {
		return errors.New("account is not active")
	}

	p := k.GetParams(ctx)
	vpnFee := p.TokenRate.MulInt64(int64(vpnGb) * int64(p.VpnGbPrice)).Int64()

	if vpnFee < 0 {
		panic("free VPN")
	}
	coins := util.Uartrs(vpnFee)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, earning.VpnCollectorName, coins); err != nil {
		return errors.Wrap(err, "cannot pay up fee")
	}
	profile.VpnLimit += uint64(vpnGb) * util.GBSize
	if err := k.SetProfile(ctx, addr, *profile); err != nil {
		return errors.Wrap(err, "cannot write profile")
	}
	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventBuyVpn{
			Address:  addr.String(),
			NewLimit: profile.VpnLimit,
			Used:     profile.VpnCurrent,
			Total:    coins.AmountOf(util.ConfigMainDenom).Uint64(),
		},
	); err != nil { panic(err) }
	return nil
}

func (k Keeper) GiveStorageUp(ctx sdk.Context, addr sdk.AccAddress, amountGb uint32) error {
	profile := k.GetProfile(ctx, addr)
	p := k.GetParams(ctx)
	newLimit := uint64(amountGb) * util.GBSize
	base := uint64(p.BaseStorageGb) * util.GBSize
	if newLimit < base {
		newLimit = base
	}
	if newLimit < profile.StorageCurrent {
		return errors.Errorf("resulting limit less than current consumption (%d < %d)", newLimit, profile.StorageCurrent)
	}
	profile.StorageLimit = newLimit
	if err := k.SetProfile(ctx, addr, *profile); err != nil {
		return errors.Wrap(err, "cannot write profile")
	}
	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventGiveUpStorage{
			Address:  addr.String(),
			NewLimit: profile.StorageLimit,
			Used:     profile.StorageCurrent,
		},
	); err != nil { panic(err) }
	return nil
}

func (k Keeper) payUpFees(ctx sdk.Context, addr sdk.AccAddress, amount sdk.Int) (int64, error) {
	fees, err := k.referralKeeper.GetReferralFeesForSubscription(ctx, addr.String())

	if err != nil {
		return 0, err
	}

	totalFee := int64(0)
	outputs := make([]bank.Output, len(fees))
	for i, fee := range fees {
		x := fee.Ratio.MulInt64(amount.Int64()).Int64()
		totalFee += x
		outputs[i] = bank.NewOutput(fee.GetBeneficiary(), util.Uartrs(x))
	}

	inputs := []bank.Input{bank.NewInput(addr, util.Uartrs(totalFee))}

	err = k.bankKeeper.InputOutputCoins(ctx, inputs, outputs)
	if err != nil {
		return totalFee, err
	}

	return totalFee, nil
}

func (k Keeper) scheduleRenew(ctx sdk.Context, addr sdk.AccAddress, time time.Time) {
	k.scheduleKeeper.ScheduleTask(ctx, time, types.RefreshHookName, addr.Bytes())
}

func (k Keeper) resetLimits(p types.Params, profile *types.Profile) {
	baseVpn := uint64(p.BaseVpnGb) * util.GBSize
	if profile.VpnCurrent > baseVpn {
		profile.VpnLimit -= profile.VpnCurrent - baseVpn
	}
	profile.VpnCurrent = 0

	if profile.StorageLimit == 0 {
		profile.StorageLimit = uint64(p.BaseStorageGb) * util.GBSize
	}
}

func (k Keeper) HandleRenewHook(ctx sdk.Context, data []byte, time time.Time) {
	if err := k.monthlyRoutine(ctx, data, time); err != nil {
		panic(err)
	}
}

func (k Keeper) monthlyRoutine(ctx sdk.Context, addr sdk.AccAddress, time time.Time) error {
	profile := k.GetProfile(ctx, addr)
	params := k.GetParams(ctx)
	if profile.IsActive(ctx) {
		// The tariff is being paid in advance
		k.scheduleRenew(ctx, addr, time.Add(k.scheduleKeeper.OneMonth(ctx)))
		k.resetLimits(params, profile)
	} else {
		// It's a payday
		if profile.AutoPay {
			if err := k.PayTariff(ctx, addr, 0); err != nil {
				defer k.referralKeeper.MustSetActive(ctx, addr.String(), false)
				if err := ctx.EventManager().EmitTypedEvents(
					&types.EventAutoPayFailed{
						Address: addr.String(),
						Error:   err.Error(),
					},
					&types.EventActivityChanged{
						Address:   addr.String(),
						ActiveNow: false,
					},
				); err != nil { panic(err) }
			}
		} else {
			defer k.referralKeeper.MustSetActive(ctx, addr.String(), false)
			if err := ctx.EventManager().EmitTypedEvent(
				&types.EventActivityChanged{
					Address:   addr.String(),
					ActiveNow: false,
				},
			); err != nil { panic(err) }
		}
	}
	if err := k.SetProfile(ctx, addr, *profile); err != nil {
		return errors.Wrap(err, "cannot write profile")
	}
	return nil
}

func (k Keeper) monthPart(ctx sdk.Context, end time.Time) util.Fraction {
	return util.FractionInt(1).Sub(util.NewFraction(end.Sub(ctx.BlockTime()).Nanoseconds(), k.scheduleKeeper.OneMonth(ctx).Nanoseconds()).Reduce())
}
