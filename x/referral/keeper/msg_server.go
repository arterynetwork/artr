package keeper

import (
	"context"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral/types"
)

type MsgServer struct {
	k Keeper
}

func NewMsgServer(k Keeper) MsgServer {
	return MsgServer{k: k}
}

var _ types.MsgServer = MsgServer{}

func (s MsgServer) RequestTransition(ctx context.Context, msg *types.MsgRequestTransition) (*types.MsgRequestTransitionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := s.k.RequestTransition(
		sdkCtx,
		msg.Subject,
		msg.Destination,
	); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgRequestTransitionResponse{}, nil
}

func (s MsgServer) ResolveTransition(ctx context.Context, msg *types.MsgResolveTransition) (*types.MsgResolveTransitionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	referrer, err := s.k.GetParent(sdkCtx, msg.Subject)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get subject's current referrer")
	}
	if msg.Signer != referrer {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "tx must be signed by the subject's current referrer")
	}
	if msg.GetApproved() {
		err = s.k.AffirmTransition(sdkCtx, msg.Subject)
	} else {
		err = s.k.CancelTransition(sdkCtx, msg.Subject, false)
	}
	if err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgResolveTransitionResponse{}, nil
}
