package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgSetLimit struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Limit   int64          `json:"limit" yaml:"limit"`
}

func NewMsgSetLimit(addr sdk.AccAddress, limit int64) MsgSetLimit {
	return MsgSetLimit{addr, limit}
}

const SetLimitConst = "set_limit"

func (msg MsgSetLimit) Route() string { return RouterKey }
func (msg MsgSetLimit) Type() string  { return SetLimitConst }
func (msg MsgSetLimit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

func (msg MsgSetLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSetLimit) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing address")
	}
	return nil
}

type MsgSetCurrent struct {
	Sender  sdk.AccAddress `json:"sender" yaml:"sender"`
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Current int64          `json:"current" yaml:"current"`
}

func NewMsgSetCurrent(sender, addr sdk.AccAddress, current int64) MsgSetCurrent {
	return MsgSetCurrent{ Sender: sender, Address: addr, Current: current }
}

const SetCurrentConst = "set_current"

func (msg MsgSetCurrent) Route() string { return RouterKey }
func (msg MsgSetCurrent) Type() string  { return SetCurrentConst }
func (msg MsgSetCurrent) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg MsgSetCurrent) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSetCurrent) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing address")
	}
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing address")
	}
	return nil
}
