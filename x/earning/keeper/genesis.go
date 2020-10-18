package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

func (k Keeper) GetEarners(ctx sdk.Context) []types.Earner {
	var result []types.Earner
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key())
		var points types.Points
		k.cdc.MustUnmarshalBinaryLengthPrefixed(it.Value(), &points)
		result = append(result, types.Earner{
			Points:  points,
			Account: acc,
		})
	}
	return result
}
