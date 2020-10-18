package earning

import (
	"github.com/arterynetwork/artr/x/earning/types"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for all the earning type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		if err := verifySignature(ctx, k, msg); err != nil { return nil, err }
		switch msg := msg.(type) {
		case types.MsgListEarners:
			return handleMsgListEarners(ctx, k, msg)
		case types.MsgRun:
			return handleMsgRun(ctx, k, msg)
		case types.MsgReset:
			return handleMsgReset(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName,  msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func verifySignature(ctx sdk.Context, k Keeper, msg sdk.Msg) error {
	ecMsg, ok := msg.(types.MsgEarningCommandI)
	if !ok { panic(fmt.Sprintf("msg supposed to have Sender")) }
	sender := ecMsg.GetSender()

	for _, signer := range k.GetParams(ctx).Signers {
		if signer.Equals(sender) { return nil }
	}
	return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "sender %s is illegal for earner txs", sender.String())
}

func handleMsgListEarners(ctx sdk.Context, k Keeper, msg types.MsgListEarners) (*sdk.Result, error) {
	err := k.ListEarners(ctx, msg.Earners)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgRun(ctx sdk.Context, k Keeper, msg types.MsgRun) (*sdk.Result, error) {
	err := k.Run(
		ctx,
		msg.FundPart,
		msg.AccountPerBlock,
		types.Points{
			Vpn:     msg.TotalVpnPoints,
			Storage: msg.TotalStoragePoints,
		},
		msg.Height,
	)
	if err != nil { return nil, err }

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgReset(ctx sdk.Context, k Keeper, _ types.MsgReset) (*sdk.Result, error) {
	k.Reset(ctx)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil

}