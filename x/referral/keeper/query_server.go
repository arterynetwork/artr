package keeper

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/referral/types"
)

type QueryServer Keeper

var _ types.QueryServer = QueryServer{}

func (qs QueryServer) Get(ctx context.Context, request *types.GetRequest) (*types.GetResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(qs)

	info, err := k.Get(sdkCtx, request.AccAddress)
	if err != nil {
		return nil, errors.Wrap(err, "cannot obtain account data")
	}

	if request.Light {
		info.Referrals = nil
		info.ActiveReferrals = nil
	}

	return &types.GetResponse{Info: info}, nil
}

func (qs QueryServer) Coins(ctx context.Context, request *types.CoinsRequest) (*types.CoinsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(qs)
	maxDepth := int(request.MaxDepth)

	//TODO: Refactor keeper to merge these calls in one
	n, err := k.GetCoinsInNetwork(sdkCtx, request.AccAddress, maxDepth)
	if err != nil {
		return nil, errors.Wrap(err, "cannot obtain total coins")
	}
	d, err := k.GetDelegatedInNetwork(sdkCtx, request.AccAddress, maxDepth)
	if err != nil {
		return nil, errors.Wrap(err, "cannot obtain delegated coins")
	}
	return &types.CoinsResponse{
		Total:     n,
		Delegated: d,
	}, nil
}

func (qs QueryServer) CheckStatus(ctx context.Context, request *types.CheckStatusRequest) (*types.CheckStatusResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(qs)

	result, err := k.AreStatusRequirementsFulfilled(sdkCtx, request.AccAddress, request.Status)
	if err != nil {
		return nil, errors.Wrap(err, "cannot obtain data")
	}
	return &types.CheckStatusResponse{Result: result}, nil
}

func (qs QueryServer) ValidateTransition(ctx context.Context, request *types.ValidateTransitionRequest) (*types.ValidateTransitionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(qs)

	if err := k.validateTransition(sdkCtx, request.Subject, request.Target, true); err == nil {
		return &types.ValidateTransitionResponse{Ok: true}, nil
	} else {
		return &types.ValidateTransitionResponse{Error: err.Error()}, nil
	}
}

func (qs QueryServer) Params(ctx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(qs)
	params := k.GetParams(sdkCtx)
	return &types.ParamsResponse{Params: params}, nil
}

func (qs QueryServer) AllWithStatus(ctx context.Context, req *types.AllWithStatusRequest) (*types.AllWithStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.Status.Validate(); err != nil {
		return nil, err
	}
	if req.Status < minIndexedStatus {
		return nil, status.Errorf(codes.NotFound, "status %s is not indexed", req.Status)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	store := sdkCtx.KVStore(qs.indexStoreKey)
	it := sdk.KVStorePrefixIterator(store, []byte{byte(req.Status)})
	defer it.Close()
	resp := types.AllWithStatusResponse{}
	for ; it.Valid(); it.Next() {
		acc := string(it.Key()[1:])
		resp.Accounts = append(resp.Accounts, acc)
	}
	return &resp, nil
}
