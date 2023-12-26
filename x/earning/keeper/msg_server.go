package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

type MsgServer Keeper

var _ types.MsgServer = MsgServer{}

func (s MsgServer) Set(ctx context.Context, msg *types.MsgSet) (resp *types.MsgSetResponse, err error) {
	_, err = s.SetMultiple(ctx, &types.MsgSetMultiple{Earners: []types.Earner{msg.Earner}})
	if err != nil {
		return nil, err
	}

	return &types.MsgSetResponse{}, nil
}

func (s MsgServer) SetMultiple(ctx context.Context, msg *types.MsgSetMultiple) (resp *types.MsgSetMultipleResponse, err error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	for _, earner := range msg.Earners {
		var acc sdk.AccAddress
		acc, err = sdk.AccAddressFromBech32(earner.Account)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "cannot parse account address: %s", earner.Account)
		}
		if earner.Vpn == nil && earner.Storage == nil {
			k.delete(sdkCtx, acc)
		} else {
			err = k.set(sdkCtx, acc, earner.GetTimestamps())
			if err != nil {
				return nil, err
			}
		}
	}

	return &types.MsgSetMultipleResponse{}, nil
}
