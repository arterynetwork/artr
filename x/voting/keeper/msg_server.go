package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/voting/types"
)

type MsgServer Keeper

var _ types.MsgServer = MsgServer{}

func (ms MsgServer) Propose(ctx context.Context, msg *types.MsgPropose) (*types.MsgProposeResponse, error) {
	var (
		sdkCtx = sdk.UnwrapSDKContext(ctx)
		k      = Keeper(ms)
	)
	if err := k.Propose(sdkCtx, *msg); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgProposeResponse{}, nil
}

func (ms MsgServer) Vote(ctx context.Context, msg *types.MsgVote) (*types.MsgVoteResponse, error) {
	var (
		sdkCtx = sdk.UnwrapSDKContext(ctx)
		k      = Keeper(ms)
	)
	if err := k.Vote(sdkCtx, msg.GetVoter(), msg.Agree); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgVoteResponse{}, nil
}
