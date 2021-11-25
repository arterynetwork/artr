package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/arterynetwork/artr/x/schedule/types"
)

type QueryServer Keeper
var _ types.QueryServer = QueryServer{}

func (qs QueryServer) All(ctx context.Context, req *types.AllRequest) (resp *types.AllResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k := Keeper(qs)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer func() {
		it.Close()

		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.All", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()

	resp = &types.AllResponse{}
	for ; it.Valid(); it.Next() {
		var sch types.Schedule
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &sch)
		resp.Tasks = append(resp.Tasks, sch.Tasks...)
	}
	return
}

func (qs QueryServer) Get(ctx context.Context, req *types.GetRequest) (resp *types.GetResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k := Keeper(qs)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	since, err := runtime.Timestamp(fmt.Sprintf(`"%s"`, req.Since))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot parse Since time")
	}
	to, err := runtime.Timestamp(fmt.Sprintf(`"%s"`, req.To))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot parse To time")
	}

	defer func(k Keeper, ctx sdk.Context) {
		if e := recover(); e != nil {
			k.Logger(ctx).Error("panic in QueryServer.Get", "req", req, "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}(k, sdkCtx)

	return &types.GetResponse{Tasks: k.GetTasks(sdkCtx, since.AsTime(), to.AsTime())}, nil
}

func (qs QueryServer) Params(ctx context.Context, req *types.ParamsRequest) (resp *types.ParamsResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	k := Keeper(qs)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	defer func(k Keeper, ctx sdk.Context) {
		if e := recover(); e != nil {
			k.Logger(ctx).Error("panic in QueryServer.Params", "req", req, "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}(k, sdkCtx)

	return &types.ParamsResponse{Params: k.GetParams(sdkCtx)}, nil
}
