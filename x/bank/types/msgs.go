package types

import (
	"github.com/golang/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RouterKey is they name of the bank module
const RouterKey = ModuleName

var _ sdk.Msg = new(MsgSend)

// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgSend(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) *MsgSend {
	return &MsgSend{
		FromAddress: fromAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
	}
}

// Route Implements Msg.
func (msg MsgSend) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSend) Type() string { return "send" }

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() error {
	if len(msg.FromAddress) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address: "+err.Error())
	}
	if len(msg.ToAddress) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing recipient address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid recipient address: "+err.Error())
	}
	coins := sdk.Coins(msg.Amount)
	if !coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}
	if !coins.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// GetSigners Implements Msg.
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
