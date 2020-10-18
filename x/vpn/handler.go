package vpn

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/vpn/types"
)

// NewHandler creates an sdk.Handler for all the vpn type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgSetLimit:
			return handleMsgSetLimit(ctx, k, msg)
		case types.MsgSetCurrent:
			return handleMsgSetCurrent(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func handleMsgSetLimit(ctx sdk.Context, k Keeper, msg types.MsgSetLimit) (*sdk.Result, error) {
	k.SetLimit(ctx, msg.Address, msg.Limit)
	return &sdk.Result{}, nil
}

func handleMsgSetCurrent(ctx sdk.Context, k Keeper, msg types.MsgSetCurrent) (*sdk.Result, error) {
	if !checkSender(ctx, k, msg.Sender) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "account %s is not in allowed sender list", msg.Sender.String())
	}
	k.SetCurrent(ctx, msg.Address, msg.Current)
	return &sdk.Result{}, nil
}

func checkSender(ctx sdk.Context, k Keeper, sender sdk.AccAddress) (allowed bool) {
	for _, signer := range k.GetParams(ctx).Signers {
		if bytes.Equal(signer, sender) {
			return true
		}
	}
	return false
}
