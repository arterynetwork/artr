package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CalculateFee(amount sdk.Int, txFee Fraction) sdk.Int {
	fee := sdk.NewInt(txFee.MulInt64(amount.Int64()).Int64())

	maxFee := sdk.NewInt(10_000000)
	if fee.GT(maxFee) {
		fee = maxFee
	}

	return fee
}

func IsSendable(denom string) bool {
	return denom == ConfigMainDenom
}
