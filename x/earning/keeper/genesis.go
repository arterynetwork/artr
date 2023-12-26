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
		var timestamps types.Timestamps
		err := k.cdc.UnmarshalBinaryBare(it.Value(), &timestamps)
		if err != nil {
			panic(err)
		}
		result = append(result, types.NewEarner(acc, timestamps.Vpn, timestamps.Storage))
	}
	return result
}

func (k Keeper) SetEarners(ctx sdk.Context, earners []types.Earner) {
	store := ctx.KVStore(k.storeKey)
	for _, earner := range earners {
		timestamps := earner.GetTimestamps()
		bz, err := k.cdc.MarshalBinaryBare(&timestamps)
		if err != nil {
			panic(err)
		}
		store.Set(earner.GetAccount(), bz)
	}
}
