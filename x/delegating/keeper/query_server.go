package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/delegating/types"
)

type QueryServer struct {
	k Keeper
}

var _ types.QueryServer = QueryServer{}

func NewQueryServer(k Keeper) QueryServer {
	return QueryServer{k: k}
}

func (q QueryServer) Params(ctx context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.k.GetParams(sdkCtx)
	return &types.ParamsResponse{Params: &params}, nil
}

func (q QueryServer) Revoking(ctx context.Context, request *types.RevokingRequest) (*types.RevokingResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	res, err := q.k.GetRevoking(sdk.UnwrapSDKContext(ctx), request.GetAccAddress())
	if err != nil {
		return nil, err
	}
	return &types.RevokingResponse{Revoking: res}, nil
}

func (q QueryServer) Accumulation(ctx context.Context, request *types.AccumulationRequest) (*types.AccumulationResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	return q.k.GetAccumulation(sdk.UnwrapSDKContext(ctx), request.GetAccAddress())
}

func (q QueryServer) Get(ctx context.Context, request *types.GetRequest) (*types.GetResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	addr, err := sdk.AccAddressFromBech32(request.AccAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot parse account address: %s", request.AccAddress)
	}
	res, err := q.k.get(sdk.UnwrapSDKContext(ctx), addr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &types.GetResponse{Data: res}, nil
}
