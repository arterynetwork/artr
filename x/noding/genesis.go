package noding

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) []abci.ValidatorUpdate {
	k.Logger(ctx).Info("Starting from genesis...")
	k.SetParams(ctx, data.Params)
	err := k.SetActiveValidators(ctx, data.Active)
	if err != nil {
		panic(err)
	}
	err = k.SetNonActiveValidators(ctx, data.NonActive)
	if err != nil {
		panic(err)
	}
	updz, err := k.GatherValidatorUpdates(ctx)
	if err != nil {
		panic(err)
	}
	return updz
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data *GenesisState) {
	params := k.GetParams(ctx)
	active, err := k.GetActiveValidators(ctx)
	if err != nil {
		panic(err)
	}
	nonactive, err := k.GetNonActiveValidators(ctx)
	if err != nil {
		panic(err)
	}
	return NewGenesisState(params, active, nonactive)
}
