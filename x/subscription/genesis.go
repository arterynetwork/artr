package subscription

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.Logger(ctx).Info("Starting from genesis...")
	k.SetParams(ctx, data.Params)
	for _, record := range data.Activity {
		k.SetActivityInfo(ctx, record.Address, record.ActivityInfo)
		if record.ActivityInfo.Active {
			if err := k.ReferralKeeper.SetActive(ctx, record.Address, true); err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	return NewGenesisState(k.GetParams(ctx), k.ExportActivity(ctx))
}
