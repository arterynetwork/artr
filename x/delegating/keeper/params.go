package keeper

import (
	"github.com/arterynetwork/artr/x/delegating/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of delegating parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.Logger(ctx).Debug("GetParams")
	k.paramspace.GetParamSet(ctx, &params)
	k.Logger(ctx).Debug("GetParams", "params", params)
	return params
}

// SetParams sets the delegating parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.Logger(ctx).Debug("SetParams", "params", params)
	k.paramspace.SetParamSet(ctx, &params)
}
