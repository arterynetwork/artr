package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank/types"
)

func (k BaseKeeper) Send(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	toAddress, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot parse recipient account address")
	}
	if k.ak.GetAccount(sdkCtx, toAddress) == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s account doesn't exist", msg.ToAddress)
	}
	if k.BlockedAddr(toAddress) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", msg.ToAddress)
	}

	for _, coin := range msg.Amount {
		if !util.IsSendable(coin.Denom) {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "tying to send forbidden denom")
		}
	}
	minCoins := sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(k.GetMinSend(sdkCtx))))
	if minCoins.IsAnyGT(msg.Amount) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "trying to send less then minimum coins")
	}

	fromAddress, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot parse sender address")
	}
	if k.BlockedSenderAddr(sdkCtx, fromAddress) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to send transactions", msg.FromAddress)
	}

	_, err = k.PayTxFee(sdkCtx, fromAddress, msg.Amount)
	if err != nil {
		logger(sdkCtx).Error(err.Error())
		return nil, err
	}
	err = k.SendCoins(sdkCtx, fromAddress, toAddress, msg.Amount)
	if err != nil {
		return nil, err
	}

	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgSendResponse{}, nil
}

func (k BaseKeeper) Burn(ctx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	account, err := sdk.AccAddressFromBech32(msg.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot parse account address")
	}

	err = k.BurnAccCoins(sdkCtx, account, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(int64(msg.Amount)))))
	if err != nil {
		return nil, err
	}

	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgBurnResponse{}, nil
}

func logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
