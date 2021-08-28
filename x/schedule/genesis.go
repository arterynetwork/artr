package schedule

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/schedule/keeper"
	"github.com/arterynetwork/artr/x/schedule/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.Logger(ctx).Info("Starting from genesis...")
	k.InitGenesis(ctx, data.Params, data.Tasks)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, tasks := k.ExportGenesis(ctx)

	return &types.GenesisState{
		Params: params,
		Tasks:  tasks,
	}
}

func DefaultGenesisState() *types.GenesisState {
	return &types.GenesisState{
		Params: types.DefaultParameters(),
	}
}
