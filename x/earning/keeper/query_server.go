package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

type QueryServer Keeper

var _ types.QueryServer = QueryServer{}

func (s QueryServer) Get(ctx context.Context, req *types.GetRequest) (resp *types.GetResponse, err error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.Get", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()

	var mResp *types.GetMultipleResponse
	mResp, err = s.GetMultiple(ctx, &types.GetMultipleRequest{Addresses: []string{req.Address}})
	if err != nil {
		return nil, err
	}

	return &types.GetResponse{Earner: mResp.Earners[0]}, nil
}

func (s QueryServer) GetMultiple(ctx context.Context, req *types.GetMultipleRequest) (resp *types.GetMultipleResponse, err error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.GetMultiple", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()

	earners := make([]types.Earner, 0)

	for _, address := range req.Addresses {
		var acc sdk.AccAddress
		acc, err = sdk.AccAddressFromBech32(address)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "cannot parse account address: %s", address)
		}

		var timestamps *types.Timestamps
		if k.has(sdkCtx, acc) {
			timestamps, err = k.get(sdkCtx, acc)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "cannot obtain earner timestamps by account address: %s", address)
			}
		} else {
			t := types.NewTimestamps(nil, nil)
			timestamps = &t
		}

		earners = append(earners, types.NewEarner(acc, timestamps.Vpn, timestamps.Storage))
	}

	return &types.GetMultipleResponse{Earners: earners}, nil
}

func (s QueryServer) List(ctx context.Context, req *types.ListRequest) (resp *types.ListResponse, err error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.List", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()

	store := sdkCtx.KVStore(k.storeKey)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	earners := make([]types.Earner, 0)
	start := req.Limit * (req.Page - 1)
	end := req.Limit * req.Page

	for current := int32(0); iterator.Valid() && (current < end); iterator.Next() {
		if current < start {
			current++
			continue
		}
		current++
		acc := sdk.AccAddress(iterator.Key())
		var timestamps types.Timestamps
		err = k.cdc.UnmarshalBinaryBare(iterator.Value(), &timestamps)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot obtain earner timestamps by account address: %s", acc.String())
		}
		earners = append(earners, types.NewEarner(acc, timestamps.Vpn, timestamps.Storage))
	}

	return &types.ListResponse{List: earners}, nil
}

func (s QueryServer) Params(ctx context.Context, _ *types.ParamsRequest) (resp *types.ParamsResponse, err error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(sdkCtx).Error("panic in QueryServer.Params", "error", e)
			err = status.Errorf(codes.Internal, "panic: %s", e)
		}
	}()
	return &types.ParamsResponse{Params: k.GetParams(sdkCtx)}, nil
}
