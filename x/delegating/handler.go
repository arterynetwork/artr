package delegating

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/delegating/types"
)

// NewHandler creates an sdk.Handler for all the delegating type messages
func NewHandler(k Keeper, supplyKeeper types.SupplyKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgDelegate:
			return handleMsgDelegate(ctx, k, supplyKeeper, msg)
		case MsgRevoke:
			return handleMsgRevoke(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func handleMsgDelegate(ctx sdk.Context, k Keeper, supplyKeeper types.SupplyKeeper, msg MsgDelegate) (*sdk.Result, error) {
	amount := msg.MicroCoins
	fee, err := util.PayTxFee(ctx, supplyKeeper, k.Logger(ctx), msg.Acc, msg.MicroCoins)
	if err != nil {
		return nil, err
	}
	amount = amount.Sub(fee)

	err = k.Delegate(ctx, msg.Acc, amount)
	if err != nil {
		k.Logger(ctx).Error(
			"cannot delegate",
			"accAddress", msg.Acc,
			"amount", amount,
			"error", err,
		)
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
