package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/noding/types"
)

type MsgServer Keeper

var _ types.MsgServer = MsgServer{}

func (s MsgServer) On(ctx context.Context, msg *types.MsgOn) (*types.MsgOnResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.SwitchOn(sdkCtx, msg.GetAccount(), msg.GetPubKey())
	if err != nil {
		return nil, err
	}
	return &types.MsgOnResponse{}, nil
}

func (s MsgServer) Off(ctx context.Context, msg *types.MsgOff) (*types.MsgOffResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.SwitchOff(sdkCtx, msg.GetAccount())
	if err != nil {
		return nil, err
	}
	return &types.MsgOffResponse{}, nil
}

func (s MsgServer) Unjail(ctx context.Context, msg *types.MsgUnjail) (*types.MsgUnjailResponse, error) {
	k := Keeper(s)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.Unjail(sdkCtx, msg.GetAccount())
	if err != nil {
		return nil, err
	}
	return &types.MsgUnjailResponse{}, nil
}
