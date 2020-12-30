package referral

import (
	"github.com/arterynetwork/artr/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	if ctx.BlockHeight()%util.BlocksOneWeek == 0 {
		if err := k.PayStatusBonus(ctx); err != nil {
			panic(err)
		}
	}
}
