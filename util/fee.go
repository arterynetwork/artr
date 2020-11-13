package util

import (
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func CalculateFee(amount sdk.Int) sdk.Int {
	fee := amount.MulRaw(3).QuoRaw(1000)

	if fee.GT(sdk.NewInt(10000000)) {
		fee = sdk.NewInt(10000000)
	}

	return fee
}

func CalculateFeeString(coins sdk.Coins) string {
	fee := CalculateFee(coins.AmountOf(ConfigMainDenom))
	return fee.String() + ConfigMainDenom
}

type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

func PayTxFee(ctx sdk.Context, k SupplyKeeper, logger log.Logger, acc sdk.AccAddress, amount sdk.Int) (fee sdk.Int, err error) {
	fee = CalculateFee(amount)
	if !fee.IsZero() {
		if err = k.SendCoinsFromAccountToModule(
			ctx, acc, auth.FeeCollectorName,
			sdk.NewCoins(sdk.NewCoin(ConfigMainDenom, fee)),
		); err != nil {
			logger.Error(
				"cannot collect fee",
				"accAddress", acc,
				"amount", amount,
				"fee", fee,
				"error", err,
			)
			return sdk.ZeroInt(), err
		}
	}
	return fee, err
}
