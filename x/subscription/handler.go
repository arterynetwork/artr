package subscription

import (
	"github.com/arterynetwork/artr/x/subscription/types"
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for all the subscription type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgPaySubscription:
			return handleMsgPaySubscription(ctx, k, msg)
		case types.MsgPayVPN:
			return handleMsgPayVPN(ctx, k, msg)
		case types.MsgPayStorage:
			return handleMsgPayStorage(ctx, k, msg)
		case types.MsgSetTokenRate:
			return handleMsgSetTokenCourse(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// handleMsgPaySubscription process payments for subscription
func handleMsgPaySubscription(ctx sdk.Context, k Keeper, msg types.MsgPaySubscription) (*sdk.Result, error) {
	err := k.PayForSubscription(ctx, msg.Address, msg.StorageAmount)

	if err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgPayVPN process payments for VPN
func handleMsgPayVPN(ctx sdk.Context, k Keeper, msg types.MsgPayVPN) (*sdk.Result, error) {
	err := k.PayForVPN(ctx, msg.Address, msg.Amount)

	if err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// MsgPayStorage process payments for storage
func handleMsgPayStorage(ctx sdk.Context, k Keeper, msg types.MsgPayStorage) (*sdk.Result, error) {
	err := k.PayForStorage(ctx, msg.Address, msg.Amount)

	if err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSetTokenCourse(ctx sdk.Context, k Keeper, msg types.MsgSetTokenRate) (*sdk.Result, error) {
	for _, signer := range k.GetParams(ctx).CourseChangeSigners {
		if !bytes.Equal(signer, msg.Sender) { continue }

		k.SetTokenCourse(ctx, msg.Value)
		return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	}
	return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "sender is not in allowed signer list")
}
