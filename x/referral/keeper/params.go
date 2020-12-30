package keeper

import (
	"github.com/arterynetwork/artr/x/referral/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of referral parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the referral parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.Logger(ctx).Debug("SetParams", "params", params)
	k.paramspace.SetParamSet(ctx, &params)
}
