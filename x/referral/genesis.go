package referral

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	//abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.Logger(ctx).Info("Starting from genesis...")
	k.SetParams(ctx, data.Params)
	for _, acc := range data.TopLevelAccounts {
		err := k.AddTopLevelAccount(ctx, acc)
		if err != nil {
			panic(err)
		}
		k.Logger(ctx).Debug("account added", "acc", acc, "parent", nil)
	}
	for _, r := range data.OtherAccounts {
		for _, acc := range r.Referrals {
			err := k.AppendChild(ctx, r.Referrer, acc)
			if err != nil {
				panic(err)
			}
			k.Logger(ctx).Debug("account added", "acc", acc, "parent", r.Referrer)
		}
	}
	if err := k.ImportFromGenesis(ctx, data.Compression, data.Downgrade, data.Transitions); err != nil {
		panic(err)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	data, err := k.ExportToGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return data
}
