package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/delegating/types"
)

type MsgServer struct {
	k Keeper
}

var _ types.MsgServer = MsgServer{}

func NewMsgServer(k Keeper) MsgServer {
	return MsgServer{k: k}
}

func (s MsgServer) Delegate(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	if err := s.k.Delegate(
		sdk.UnwrapSDKContext(ctx),
		msg.GetAddress(),
		msg.MicroCoins,
	); err != nil {
		return nil, err
	}
	return &types.MsgDelegateResponse{}, nil
}

func (s MsgServer) Revoke(ctx context.Context, msg *types.MsgRevoke) (*types.MsgRevokeResponse, error) {
	if err := s.k.Revoke(
		sdk.UnwrapSDKContext(ctx),
		msg.GetAddress(),
		msg.MicroCoins,
	); err != nil {
		return nil, err
	}
	return &types.MsgRevokeResponse{}, nil
}
