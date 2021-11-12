package referral

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.Logger(ctx).Info("Starting from genesis...")
	k.SetParams(ctx, data.Params)

	if err := k.ImportFromGenesis(
		ctx,
		data.TopLevelAccounts,
		data.OtherAccounts,
		data.BanishedAccounts,
		data.Compressions,
		data.Banishment,
		data.Downgrades,
		data.Transitions,
	); err != nil {
		panic(err)
	}
	k.Logger(ctx).Info("... all done!")
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data *GenesisState) {
	data, err := k.ExportToGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return data
}
