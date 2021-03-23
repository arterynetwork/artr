package bank

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank/internal/keeper"
	"github.com/arterynetwork/artr/x/bank/internal/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k keeper.Keeper, sk types.SupplyKeeper, ak types.AccountKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSend:
			return handleMsgSend(ctx, k, sk, ak, msg)

		case types.MsgMultiSend:
			return handleMsgMultiSend(ctx, k, sk, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized bank message type: %T", msg)
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, k keeper.Keeper, sk types.SupplyKeeper, ak types.AccountKeeper, msg types.MsgSend) (*sdk.Result, error) {
	if !k.GetSendEnabled(ctx) {
		return nil, types.ErrSendDisabled
	}

	if ak.GetAccount(ctx, msg.ToAddress) == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s account doesn't exist", msg.ToAddress)
	}

	if k.BlacklistedAddr(msg.ToAddress) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", msg.ToAddress)
	}

	minCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(k.GetMinSend(ctx))))

	if minCoins.IsAnyGT(msg.Amount) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "tying to send less then minimum coins")
	}

	for _, coin := range msg.Amount {
		if strings.ToLower(coin.Denom) != strings.ToLower(util.ConfigMainDenom) {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "tying to send forbidden denom")
		}
	}

	amount := msg.Amount.AmountOf(util.ConfigMainDenom)
	_, err := util.PayTxFee(ctx, sk, logger(ctx), msg.FromAddress, amount)
	if err != nil {
		return nil, err
	}
	err = k.SendCoins(ctx, msg.FromAddress, msg.ToAddress, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, amount)))
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// Handle MsgMultiSend.
func handleMsgMultiSend(ctx sdk.Context, k keeper.Keeper, sk types.SupplyKeeper, msg types.MsgMultiSend) (*sdk.Result, error) {
	// NOTE: totalIn == totalOut should already have been checked
	if !k.GetSendEnabled(ctx) {
		return nil, types.ErrSendDisabled
	}

	for _, in := range msg.Inputs {
		for _, coin := range in.Coins {
			if strings.ToLower(coin.Denom) != strings.ToLower(util.ConfigMainDenom) {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "tying to send forbidden denom")
			}
		}
		_, err := util.PayTxFee(ctx, sk, logger(ctx), in.Address, in.Coins.AmountOf(util.ConfigMainDenom))
		if err != nil {
			return nil, err
		}
	}

	for _, out := range msg.Outputs {
		if k.BlacklistedAddr(out.Address) {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", out.Address)
		}
	}

	err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
