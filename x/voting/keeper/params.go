package keeper

import (
	"github.com/arterynetwork/artr/x/voting/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of voting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the voting parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramspace.SetParamSet(ctx, &params)
}
