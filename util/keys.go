package util

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ConfigMainDenom      = "uartr"
	ConfigDelegatedDenom = "uartrd"
	ConfigRevokingDenom  = "uartrr"
	ConfigStakeDenom     = "stake"
	GBSize               = 1024 * 1024 * 1024
	BlocksOneDay         = 2880
	BlocksOneWeek        = BlocksOneDay * 7
	BlocksOneMonth       = BlocksOneDay * 30
	BlocksOneHour        = BlocksOneDay / 24
)

func Uartrs(n int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(ConfigMainDenom, sdk.NewInt(n)))
}
