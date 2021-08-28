package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/profile/types"
)

// GetParams returns the total set of subscription parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the subscription parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.Logger(ctx).Debug("SetParams", "params", params)
	if k.paramspace.Has(ctx, types.KeyCardMagic) {
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
	params.Creators = add(params.Creators, creator.String())
	k.SetParams(ctx, params)
}

func (k Keeper) RemoveFreeCreator(ctx sdk.Context, creator sdk.AccAddress) {
	params := k.GetParams(ctx)
	params.Creators = remove(params.Creators, creator.String())
	k.SetParams(ctx, params)
}

func (k Keeper) AddTokenRateSigner(ctx sdk.Context, signer sdk.AccAddress) {
	params := k.GetParams(ctx)
	params.TokenRateSigners = add(params.TokenRateSigners, signer.String())
	k.SetParams(ctx, params)
}

func (k Keeper) RemoveTokenRateSigner(ctx sdk.Context, signer sdk.AccAddress) {
	params := k.GetParams(ctx)
	params.TokenRateSigners = remove(params.TokenRateSigners, signer.String())
	k.SetParams(ctx, params)
}

func (k Keeper) AddVpnCurrentSigner(ctx sdk.Context, signer sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.VpnSigners = add(p.VpnSigners, signer.String())
	k.SetParams(ctx, p)
}

func (k Keeper) RemoveVpnCurrentSigner(ctx sdk.Context, signer sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.VpnSigners = remove(p.VpnSigners, signer.String())
	k.SetParams(ctx, p)
}

func (k Keeper) AddStorageCurrentSigner(ctx sdk.Context, signer sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.StorageSigners = add(p.StorageSigners, signer.String())
	k.SetParams(ctx, p)
}

func (k Keeper) RemoveStorageCurrentSigner(ctx sdk.Context, signer sdk.AccAddress) {
	p := k.GetParams(ctx)
	p.StorageSigners = remove(p.StorageSigners, signer.String())
	k.SetParams(ctx, p)
}

func add(arr []string, item string) []string {
	for _, x := range arr {
		if x == item {
			return arr
		}
	}
	return append(arr, item)
}

func remove(arr []string, item string) []string {
	idx := -1
	for i, x := range arr {
		if x == item {
			idx = i
			break
		}
	}
	if idx < 0 {
		return arr
	}
	if idx != len(arr)-1 {
		arr[idx] = arr[len(arr)-1]
	}
	return arr[:len(arr)-1]
}
