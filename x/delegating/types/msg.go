package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// verify interface at compile time
var _ sdk.Msg = &MsgDelegate{}

// MsgDelegate - struct for delegating coins
type MsgDelegate struct {
	Acc sdk.AccAddress `json:"address" yaml:"address"`
	MicroCoins sdk.Int `json:"micro_coins" yaml:"micro_coins"`
}

// NewMsgDelegate creates a new MsgDelegate instance
func NewMsgDelegate(acc sdk.AccAddress, ucoins sdk.Int) MsgDelegate {
	return MsgDelegate{
		Acc: acc,
		MicroCoins: ucoins,
	}
}

const DelegateConst = "delegate"

// nolint
func (msg MsgDelegate) Route() string { return RouterKey }
func (msg MsgDelegate) Type() string  { return DelegateConst }
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Acc}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgDelegate) ValidateBasic() error {
	if msg.Acc.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing account address")
	}
	if !msg.MicroCoins.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "amount must be positive")
	}
	return nil
}

// verify interface at compile time
var _ sdk.Msg = &MsgRevoke{}

// MsgRevoke - struct for revoking coins from delegating
type MsgRevoke struct {
	Acc sdk.AccAddress `json:"address" yaml:"address"`
	MicroCoins sdk.Int `json:"micro_coins"`
}

// NewMsgDelegate creates a new MsgDelegate instance
func NewMsgRevoke(acc sdk.AccAddress, ucoins sdk.Int) MsgRevoke {
	return MsgRevoke{
		Acc: acc,
		MicroCoins: ucoins,
	}
}

const RevokeConst = "revoke"

// nolint
func (msg MsgRevoke) Route() string { return RouterKey }
func (msg MsgRevoke) Type() string  { return RevokeConst }
func (msg MsgRevoke) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Acc}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgRevoke) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgRevoke) ValidateBasic() error {
	if msg.Acc.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing account address")
	}
	if !msg.MicroCoins.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "amount must be positive")
	}
	return nil
}
