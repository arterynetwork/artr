package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

type MsgServer Keeper

var _ types.MsgServer = MsgServer{}

func (s MsgServer) ListEarners(ctx context.Context, msg *types.MsgListEarners) (*types.MsgListEarnersResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := k.ListEarners(sdkCtx, msg.Earners); err != nil {
		return nil, err
	}
	return &types.MsgListEarnersResponse{}, nil
}

func (s MsgServer) Run(ctx context.Context, msg *types.MsgRun) (*types.MsgRunResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := k.Run(
		sdkCtx,
		msg.FundPart,
		msg.PerBlock,
		types.Points{
			Vpn:     msg.TotalVpn,
			Storage: msg.TotalStorage,
		},
		msg.Time,
	); err != nil {
		return nil, err
	}
	return &types.MsgRunResponse{}, nil
}

func (s MsgServer) Reset(ctx context.Context, msg *types.MsgReset) (*types.MsgResetResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Reset(sdkCtx)
	return &types.MsgResetResponse{}, nil
}
