package noding

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for all the noding type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgSwitchOn:
			return handleMsgSwitchOn(ctx, k, msg)
		case MsgSwitchOff:
			return handleMsgSwitchOff(ctx, k, msg)
		case MsgUnjail:
			return handleMsgUnjail(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func handleMsgSwitchOn(ctx sdk.Context, k Keeper, msg MsgSwitchOn) (*sdk.Result, error) {
	err := k.SwitchOn(ctx, msg.AccAddress, msg.PubKey)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}

func handleMsgSwitchOff(ctx sdk.Context, k Keeper, msg MsgSwitchOff) (*sdk.Result, error) {
	err := k.SwitchOff(ctx, msg.AccAddress)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}

func handleMsgUnjail(ctx sdk.Context, k Keeper, msg MsgUnjail) (*sdk.Result, error) {
	err := k.Unjail(ctx, msg.AccAddress)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}
