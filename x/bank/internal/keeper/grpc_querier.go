package keeper

import (
	"github.com/arterynetwork/artr/x/bank/types"
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueryServer BaseKeeper

var _ types.QueryServer = BaseKeeper{}

func (k BaseKeeper) Params(ctx context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params := k.GetParams(sdkCtx)

	return &types.ParamsResponse{
		Params: params,
	}, nil
}

func (k BaseKeeper) Supply(ctx context.Context, req *types.SupplyRequest) (*types.SupplyResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	supply := k.GetSupply(sdkCtx)

	return &types.SupplyResponse{Supply: supply}, nil
}

func (k BaseKeeper) Balance(ctx context.Context, req *types.BalanceRequest) (*types.BalanceResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	addr, err := sdk.AccAddressFromBech32(req.AccAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot parse account address")
	}

	balance := k.GetBalance(sdkCtx, addr)

	return &types.BalanceResponse{Balance: balance}, nil
}
