package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CalculateFee(amount sdk.Int, txFeeFraction Fraction, txFeeMaxAmount int64) sdk.Int {
	fee := sdk.NewInt(txFeeFraction.MulInt64(amount.Int64()).Int64())

	maxFee := sdk.NewInt(txFeeMaxAmount)
	if fee.GT(maxFee) {
		fee = maxFee
	}

	return fee
}

func IsSendable(denom string) bool {
	return denom == ConfigMainDenom
}
