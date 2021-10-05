package profile

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/profile/types"
)

// InitGenesis initialize default parameters
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.Logger(ctx).Info("Starting from genesis...")
	k.SetParams(ctx, data.Params)
	k.ImportProfileRecords(ctx, data.Profiles)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) (data *types.GenesisState) {
	return types.NewGenesisState(k.GetParams(ctx), k.ExportProfileRecords(ctx))
}
