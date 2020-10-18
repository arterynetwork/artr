package storage

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, limit := range data.Limits {
		k.SetLimit(ctx, limit.Account, int64(limit.Volume))
	}
	for _, current := range data.Current {
		k.SetCurrent(ctx, current.Account, int64(current.Volume))
	}
	for i, d := range data.Data {
		bz, err := base64.StdEncoding.DecodeString(d.Base64)
		if err != nil { panic(sdkerrors.Wrapf(err, "malformed base64 (data.#%d)", i))}
		k.SetData(ctx, d.Account, bz)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	return NewGenesisState(k.ExportLimits(ctx), k.ExportCurrent(ctx), k.ExportData(ctx))
}
