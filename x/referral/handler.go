package referral

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/referral/types"
)

// NewHandler creates a sdk.Handler for all the referral module's messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgRequestTransition:
			return handleMsgRequestTransition(ctx, k, msg)
		case types.MsgResolveTransition:
			return handleMsgResolveTransition(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"unrecognized %s message type: %T", ModuleName, msg,
			)
		}
	}
}

func handleMsgRequestTransition(ctx sdk.Context, k Keeper, msg types.MsgRequestTransition) (*sdk.Result, error) {
	err := k.RequestTransition(ctx, msg.Subject, msg.Destination)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgResolveTransition(ctx sdk.Context, k Keeper, msg types.MsgResolveTransition) (*sdk.Result, error) {
	referrer, err := k.GetParent(ctx, msg.Subject)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get subject's current referrer")
	}
	if !referrer.Equals(msg.Sender) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "tx must be signed by the subject's current referrer")
	}
	if msg.Approved {
		err = k.AffirmTransition(ctx, msg.Subject)
	} else {
		err = k.CancelTransition(ctx, msg.Subject, false)
	}
	if err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
