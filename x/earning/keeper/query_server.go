package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

type QueryServer Keeper

var _ types.QueryServer = QueryServer{}

func (s QueryServer) List(ctx context.Context, _ *types.ListRequest) (*types.ListResponse, error) {
	return &types.ListResponse{List: Keeper(s).GetEarners(sdk.UnwrapSDKContext(ctx))}, nil
}

func (s QueryServer) State(ctx context.Context, _ *types.StateRequest) (*types.StateResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.StateResponse{State: k.GetState(sdkCtx)}, nil
}

func (s QueryServer) Params(ctx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.ParamsResponse{Params: k.GetParams(sdkCtx)}, nil
}
