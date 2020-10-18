package delegating

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.SetParams(ctx, data.Params)
	k.InitClusters(ctx, data.Clusters)
	k.InitRevokeRequests(ctx, data.Revoking)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	return NewGenesisState(
		k.GetParams(ctx),
		k.ExportClusters(ctx),
		k.ExportRevokeRequests(ctx),
	)
}
