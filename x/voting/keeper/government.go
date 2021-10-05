package keeper

import (
	"github.com/arterynetwork/artr/x/voting/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/proto"
)

func (k Keeper) GetGovernment(ctx sdk.Context) types.Government {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyGovernment)

	var gov types.Government

	err := proto.Unmarshal(bz, &gov)
	if err != nil {
		panic(err)
	}

	return gov
}

func (k Keeper) SetGovernment(ctx sdk.Context, gov types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&gov)
	if err != nil {
		panic(err)
	}
	store.Set(types.KeyGovernment, bz)
}

func (k Keeper) RemoveGovernor(ctx sdk.Context, gov sdk.AccAddress) {
	govs := k.GetGovernment(ctx)
	govs.Remove(gov)
	k.SetGovernment(ctx, govs)
}

func (k Keeper) AddGovernor(ctx sdk.Context, gov sdk.AccAddress) {
	govs := k.GetGovernment(ctx)
	govs.Append(gov)
	k.SetGovernment(ctx, govs)
}
