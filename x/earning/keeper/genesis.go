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
		err := k.cdc.UnmarshalBinaryBare(it.Value(), &points)
		if err != nil {
			panic(err)
		}
		result = append(result, types.NewEarner(acc, points.Vpn, points.Storage))
	}
	return result
}
