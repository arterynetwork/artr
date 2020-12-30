package earning

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.Logger(ctx).Info("Starting from genesis...")
	k.SetParams(ctx, data.Params)
	k.SetState(ctx, types.NewStateUnlocked())
	if err := k.ListEarners(ctx, data.Earners); err != nil {
		panic(err)
	}
	k.SetState(ctx, data.State)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return NewGenesisState(
		k.GetParams(ctx),
		k.GetState(ctx),
		k.GetEarners(ctx),
	)
}
