package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/vpn/types"
)

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramspace.SetParamSet(ctx, &params)
}

func (k Keeper) AddSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.Signers = append(p.Signers, address)
	k.SetParams(ctx, p)
}

func (k Keeper) RemoveSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	for i, s := range p.Signers {
		if bytes.Equal(s, address) {
			last := len(p.Signers) - 1
			if i != last {
				p.Signers[i] = p.Signers[last]
			}
			p.Signers = p.Signers[:last]
			k.SetParams(ctx, p)
			break
		}
	}
}
