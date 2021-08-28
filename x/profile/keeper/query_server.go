package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/profile/types"
)

type QueryServer Keeper

var _ types.QueryServer = QueryServer{}

func (q QueryServer) GetByAddress(ctx context.Context, req *types.GetByAddressRequest) (*types.GetByAddressResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	profile := Keeper(q).GetProfile(sdkCtx, req.GetAddress())
	if profile == nil {
		return nil, types.ErrNotFound
	}
	return &types.GetByAddressResponse{
		Profile: *profile,
	}, nil
}

func (q QueryServer) GetByNickname(ctx context.Context, req *types.GetByNicknameRequest) (*types.GetByNicknameResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	addr := Keeper(q).GetProfileAccountByNickname(sdkCtx, req.Nickname)
	if addr.Empty() {
		return nil, types.ErrNotFound
	}
	profile := Keeper(q).GetProfile(sdkCtx, addr)
	return &types.GetByNicknameResponse{
		Address: addr.String(),
		Profile: *profile,
	}, nil
}

func (q QueryServer) GetByCardNumber(ctx context.Context, req *types.GetByCardNumberRequest) (*types.GetByCardNumberResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	addr := Keeper(q).GetProfileAccountByCardNumber(sdkCtx, req.CardNumber)
	if addr.Empty() {
		return nil, types.ErrNotFound
	}
	profile := Keeper(q).GetProfile(sdkCtx, addr)
	return &types.GetByCardNumberResponse{
		Address: addr.String(),
		Profile: *profile,
	}, nil
}

func (q QueryServer) Params(ctx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := Keeper(q).GetParams(sdkCtx)
	return &types.ParamsResponse{
		Params: params,
	}, nil
}
