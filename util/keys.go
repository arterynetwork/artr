package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ConfigMainDenom      = "uartr"
	ConfigDelegatedDenom = "uartrd"
	ConfigRevokingDenom  = "uartrr"

	SplittableFeeCollectorName      = "splittable_fee_collector"
	TransactionFeeSplitRatiosMaxLcm = 100

	GBSize         = 1024 * 1024 * 1024
	BlocksOneDay   = 2880
	BlocksOneWeek  = BlocksOneDay * 7
	BlocksOneMonth = BlocksOneDay * 30
	BlocksOneHour  = BlocksOneDay / 24
)

func Uartrs(n int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(ConfigMainDenom, sdk.NewInt(n)))
}

func UartrsUint64(n uint64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(ConfigMainDenom, sdk.NewIntFromUint64(n)))
}
