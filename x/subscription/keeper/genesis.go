package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/arterynetwork/artr/x/subscription/types"
)

func (k Keeper) ExportActivity(ctx sdk.Context) []types.GenesisActivityInfo {
	var result []types.GenesisActivityInfo
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key()[len(auth.AddressStoreKeyPrefix):])
		var info types.ActivityInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(it.Value(), &info)
		result = append(result, types.GenesisActivityInfo{
			Address:      acc,
			ActivityInfo: info,
		})
	}
	return result
}
