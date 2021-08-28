package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/voting/types"
)

type QueryServer Keeper

var _ types.QueryServer = QueryServer{}

func (qs QueryServer) History(ctx context.Context, req *types.HistoryRequest) (*types.HistoryResponse, error) {
	var (
		sdkCtx = sdk.UnwrapSDKContext(ctx)
		k      = Keeper(qs)
	)
	data := k.GetHistory(sdkCtx, req.Limit, req.Page)
	return &types.HistoryResponse{
		History: data,
	}, nil
}

func (qs QueryServer) Government(ctx context.Context, _ *types.GovernmentRequest) (*types.GovernmentResponse, error) {
	var (
		sdkCtx = sdk.UnwrapSDKContext(ctx)
		k      = Keeper(qs)
	)
	data := k.GetGovernment(sdkCtx)
	return &types.GovernmentResponse{
		Members: data.Members,
	}, nil
}

func (qs QueryServer) Current(ctx context.Context, _ *types.CurrentRequest) (*types.CurrentResponse, error) {
	var (
		sdkCtx = sdk.UnwrapSDKContext(ctx)
		k      = Keeper(qs)
	)
	var (
		proposal = k.GetCurrentProposal(sdkCtx)
		gov,
		agreed,
		disagreed types.Government
	)
	if proposal == nil {
		proposal = new(types.Proposal)
	} else {
		gov = k.GetGovernment(sdkCtx)
		agreed = k.GetAgreed(sdkCtx)
		disagreed = k.GetDisagreed(sdkCtx)
	}
	return &types.CurrentResponse{
		Proposal:   *proposal,
		Government: gov.Strings(),
		Agreed:     agreed.Strings(),
		Disagreed:  disagreed.Strings(),
	}, nil
}

func (qs QueryServer) Params(ctx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	var (
		sdkCtx = sdk.UnwrapSDKContext(ctx)
		k      = Keeper(qs)
	)
	data := k.GetParams(sdkCtx)
	return &types.ParamsResponse{
		Params: data,
	}, nil
}
