package keeper

import (
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

func (k Keeper) AddSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.Signers = append(p.Signers, address.String())
	k.SetParams(ctx, p)
}

func (k Keeper) RemoveSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	bech32 := address.String()
	for i, signer := range p.Signers {
		if signer == bech32 {
			last := len(p.Signers) - 1
			if i != last {
				p.Signers[i] = p.Signers[last]
			}
			p.Signers = p.Signers[:last]
			k.SetParams(ctx, p)
			return
		}
	}
}
