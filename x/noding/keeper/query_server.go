package keeper

import (
	"context"
	"encoding/binary"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/noding/types"
)

type QueryServer Keeper

var _ types.QueryServer = QueryServer{}

func (s QueryServer) Params(ctx context.Context, _ *types.ParamsRequest) (resp *types.ParamsResponse, err error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.Params", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	params := k.GetParams(sdkCtx)
	return &types.ParamsResponse{
		Params: params,
	}, nil
}

func (s QueryServer) Get(ctx context.Context, req *types.GetRequest) (resp *types.GetResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.Get", "error", e, "request", *req)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	info, err := k.Get(sdkCtx, req.GetAccount())
	if err != nil {
		return nil, err
	}

	return &types.GetResponse{
		Info: info,
	}, nil
}

func (s QueryServer) Proposer(ctx context.Context, req *types.ProposerRequest) (resp *types.ProposerResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.Proposer", "error", e, "request", *req)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	proposer, err := k.GetBlockProposer(sdkCtx, req.Height)
	if err != nil {
		return nil, err
	}
	return &types.ProposerResponse{
		Account: proposer.String(),
	}, nil
}

func (s QueryServer) IsAllowed(ctx context.Context, req *types.IsAllowedRequest) (resp *types.IsAllowedResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.IsAllowed", "error", e, "request", *req)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	verdict, _, reason, err := k.IsQualified(sdkCtx, req.GetAccount())
	if err != nil {
		return nil, err
	}
	return &types.IsAllowedResponse{
		Verdict: verdict,
		Reason:  reason,
	}, nil
}

func (s QueryServer) Operator(ctx context.Context, req *types.OperatorRequest) (resp *types.OperatorResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.Operator", "error", e, "request", *req)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	acc, found := k.getNodeOperatorFromIndex(sdkCtx, req.GetConsAddress())
	if !found {
		return nil, types.ErrNotFound
	}
	return &types.OperatorResponse{
		Account: acc.String(),
	}, nil
}

func (s QueryServer) SwitchedOn(ctx context.Context, _ *types.SwitchedOnRequest) (resp *types.SwitchedOnResponse, err error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.SwitchedOn", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	list, err := k.GetActiveValidatorList(sdkCtx)
	if err != nil {
		return nil, err
	}
	return types.NewSwitchedOnResponse(list), err
}

func (s QueryServer) Queue(ctx context.Context, _ *types.QueueRequest) (resp *types.QueueResponse, err error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.indexStoreKey)
	resp = &types.QueueResponse{}
	it := sdk.KVStorePrefixIterator(store, IdxPrefixLotteryQueue)
	defer func() {
		it.Close()

		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.Queue", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	for ; it.Valid(); it.Next() {
		var (
			no  uint64         = binary.BigEndian.Uint64(it.Key())
			acc sdk.AccAddress = it.Value()
		)
		resp.Queue = append(resp.Queue, types.QueueResponse_Validator{No: no, Account: acc.String()})
	}
	return
}

func (s QueryServer) State(ctx context.Context, req *types.StateRequest) (resp *types.StateResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.State", "error", e, "request", *req)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	addr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot parse account address: %s", req.Account)
	}
	return &types.StateResponse{State: k.GetValidatorState(sdkCtx, addr)}, nil
}
