package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/profile/types"
)

func (k Keeper) ExportProfileRecords(ctx sdk.Context) []types.GenesisProfile {
	var result []types.GenesisProfile
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key()[1:])
		var value types.Profile
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &value)
		value.CardNumber = 0
		result = append(result, types.GenesisProfile{
			Address: acc,
			Profile: value,
		})
	}
	return result
}
