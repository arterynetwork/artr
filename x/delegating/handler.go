package delegating

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for all the delegating type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgDelegate:
			return handleMsgDelegate(ctx, k, msg)
		case MsgRevoke:
			return handleMsgRevoke(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName,  msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func handleMsgDelegate(ctx sdk.Context, k Keeper, msg MsgDelegate) (*sdk.Result, error) {
	err := k.Delegate(ctx, msg.Acc, msg.MicroCoins)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgRevoke(ctx sdk.Context, k Keeper, msg MsgRevoke) (*sdk.Result, error) {
	err := k.Revoke(ctx, msg.Acc, msg.MicroCoins)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
