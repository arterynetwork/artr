package keeper

import (
	"github.com/arterynetwork/artr/x/schedule/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of noding parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the noding parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramspace.SetParamSet(ctx, &params)
}
