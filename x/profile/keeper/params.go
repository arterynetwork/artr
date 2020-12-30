package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/arterynetwork/artr/x/profile/types"
)

// GetParams returns the total set of subscription parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the subscription parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if k.paramspace.(subspace.Subspace).Has(ctx, types.KeyCardMagic) {
		var oldCardMagic uint64
		k.paramspace.Get(ctx, types.KeyCardMagic, &oldCardMagic)

		if params.CardMagic != oldCardMagic {
			if params.CardMagic == 0 {
				params.CardMagic = oldCardMagic
			} else {
				panic("card number magic must not be changed")
			}
		}
	}
	k.paramspace.SetParamSet(ctx, &params)
}

func (k Keeper) AddFreeCreator(ctx sdk.Context, creator sdk.AccAddress) {
	params := k.GetParams(ctx)
	for _, c := range params.Creators {
		if c.Equals(creator) {
			return
		}
	}
	params.Creators = append(params.Creators, creator)
	k.SetParams(ctx, params)
}

func (k Keeper) RemoveFreeCreator(ctx sdk.Context, creator sdk.AccAddress) {
	params := k.GetParams(ctx)
	idx := -1
	for i, c := range params.Creators {
		if c.Equals(creator) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	if idx != len(params.Creators)-1 {
		params.Creators[idx] = params.Creators[len(params.Creators)-1]
	}
	params.Creators = params.Creators[:len(params.Creators)-1]
	k.SetParams(ctx, params)
}
