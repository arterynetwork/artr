package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
