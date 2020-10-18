package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/subscription/types"
)

// GetParams returns the total set of subscription parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the subscription parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramspace.SetParamSet(ctx, &params)
}

func (k Keeper) SetTokenCourse(ctx sdk.Context, value uint32) {
	p := k.GetParams(ctx)
	p.TokenCourse = value
	k.SetParams(ctx, p)
}

func (k Keeper) AddCourseChangeSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.CourseChangeSigners = append(p.CourseChangeSigners, address)
	k.SetParams(ctx, p)
}

func (k Keeper) RemoveCourseChangeSigner(ctx sdk.Context, address sdk.AccAddress) {
	p := k.GetParams(ctx)
	for i, signer := range p.CourseChangeSigners {
		if bytes.Equal(signer.Bytes(), address.Bytes()) {
			last := len(p.CourseChangeSigners) - 1
			if i != last {
				p.CourseChangeSigners[i] = p.CourseChangeSigners[last]
				p.CourseChangeSigners = p.CourseChangeSigners[:last]
				k.SetParams(ctx, p)
				return
			}
		}
	}
}
