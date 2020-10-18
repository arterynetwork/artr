package keeper

import (
	"github.com/arterynetwork/artr/x/voting/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetGovernment(ctx sdk.Context) types.Government {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyGovernment)

	var gov types.Government

	k.cdc.MustUnmarshalBinaryBare(bz, &gov)

	return gov
}

func (k Keeper) SetGovernment(ctx sdk.Context, gov types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(gov)
	store.Set(types.KeyGovernment, bz)
}

func (k Keeper) RemoveGovernor(ctx sdk.Context, gov sdk.AccAddress) {
	govs := k.GetGovernment(ctx)
	k.SetGovernment(ctx, govs.Remove(gov))
}

func (k Keeper) AddGovernor(ctx sdk.Context, gov sdk.AccAddress) {
	govs := k.GetGovernment(ctx)
	k.SetGovernment(ctx, govs.Append(gov))
}
