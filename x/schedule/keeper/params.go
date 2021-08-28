package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/schedule/types"
)

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) setParams(ctx sdk.Context, params types.Params) {
	k.paramspace.SetParamSet(ctx, &params)
}
