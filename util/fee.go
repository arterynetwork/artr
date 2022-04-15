package util

import (
	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func CalculateFee(amount sdk.Int) sdk.Int {
	fee := amount.MulRaw(3).QuoRaw(1000)

	maxFee := sdk.NewInt(10_000000)
	if fee.GT(maxFee) {
		fee = maxFee
	}

	return fee
}

func IsSendable(denom string) bool {
	return denom == ConfigMainDenom
}

type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

func PayTxFee(ctx sdk.Context, k SupplyKeeper, logger log.Logger, acc sdk.AccAddress, amount sdk.Coins) (fee sdk.Coins, err error) {
	for _, c := range amount {
		if !IsSendable(c.Denom) { panic(errors.Errorf("%s is not sendable", c.Denom)) }
		fee = fee.Add(sdk.NewCoin(c.Denom, CalculateFee(c.Amount)))
	}
	if !fee.IsZero() {
		if err = k.SendCoinsFromAccountToModule(
			ctx, acc, auth.FeeCollectorName,
			fee,
		); err != nil {
			logger.Error(
				"cannot collect fee",
				"accAddress", acc,
				"amount", amount,
				"fee", fee,
				"error", err,
			)
			return fee, err
		}
	}
	return fee, err
}
