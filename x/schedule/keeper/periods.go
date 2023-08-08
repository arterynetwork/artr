package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) OneDay(ctx sdk.Context) time.Duration {
	return time.Duration(k.GetParams(ctx).DayNanos)
}
func (k Keeper) OneWeek(ctx sdk.Context) time.Duration  { return 7 * k.OneDay(ctx) }
func (k Keeper) OneMonth(ctx sdk.Context) time.Duration { return 30 * k.OneDay(ctx) }
