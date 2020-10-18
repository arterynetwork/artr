package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

// GetParams returns the total set of earning parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the earning parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.Logger(ctx).Debug("SetParams", "params", params)
	k.paramspace.SetParamSet(ctx, &params)
}

func (k Keeper) GetState(ctx sdk.Context) (state types.StateParams) {
	k.paramspace.GetParamSet(ctx, &state)
	return state
}

func (k Keeper) SetState(ctx sdk.Context, state types.StateParams) {
	k.Logger(ctx).Debug("SetState", "state", state)
	k.paramspace.SetParamSet(ctx, &state)
}

func (k Keeper) AddSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.Signers = append(p.Signers, address)
	k.SetParams(ctx, p)
}

func (k Keeper) RemoveSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	for i, signer := range p.Signers {
		if bytes.Equal(signer, address) {
			last := len(p.Signers) -1
			if i != last {
				p.Signers[i] = p.Signers[last]
			}
			p.Signers = p.Signers[:last]
			k.SetParams(ctx, p)
			return
		}
	}
}
