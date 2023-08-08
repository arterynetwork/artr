package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CalculateFee(amount sdk.Int, txFeeFraction Fraction, txFeeMaxAmount int64, forProposerFeeFraction, forCompanyFeeFraction Fraction) sdk.Int {
	fee := sdk.NewInt(txFeeFraction.MulInt64(amount.Int64()).Int64())

	maxFee := sdk.NewInt(txFeeMaxAmount)
	if !maxFee.IsZero() && fee.GT(maxFee) {
		fee = maxFee
	}

	return calculateSplittableFee(fee, forProposerFeeFraction, forCompanyFeeFraction)
}

func calculateForBurningFeeFraction(forProposerFeeFraction, forCompanyFeeFraction Fraction) Fraction {
	return FractionInt(1).Sub(forProposerFeeFraction).Sub(forCompanyFeeFraction)
}

func CalculateTransactionFeeSplitRatiosLCM(forProposerFeeFraction, forCompanyFeeFraction Fraction) sdk.Int {
	return sdk.NewIntFromBigInt(lcm(lcm(forProposerFeeFraction.denom, forCompanyFeeFraction.denom), calculateForBurningFeeFraction(forProposerFeeFraction, forCompanyFeeFraction).denom))
}

func calculateSplittableFee(feeLimit sdk.Int, forProposerFeeFraction, forCompanyFeeFraction Fraction) sdk.Int {
	return feeLimit.Sub(feeLimit.Mod(CalculateTransactionFeeSplitRatiosLCM(forProposerFeeFraction, forCompanyFeeFraction)))
}

func SplitFee(splittableFee sdk.Int, forProposerFeeFraction, forCompanyFeeFraction Fraction) (forProposer, forCompany, forBurning sdk.Int) {
	return sdk.NewInt(forProposerFeeFraction.MulInt64(splittableFee.Int64()).Int64()),
		sdk.NewInt(forCompanyFeeFraction.MulInt64(splittableFee.Int64()).Int64()),
		sdk.NewInt(calculateForBurningFeeFraction(forProposerFeeFraction, forCompanyFeeFraction).MulInt64(splittableFee.Int64()).Int64())
}

func IsSendable(denom string) bool {
	return denom == ConfigMainDenom
}
