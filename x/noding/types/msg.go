package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto"
)

// verify interface at compile time
var (
	_ sdk.Msg = &MsgSwitchOn{}
	_ sdk.Msg = &MsgSwitchOff{}
	_ sdk.Msg = &MsgUnjail{}
)

type MsgSwitchOn struct {
	AccAddress sdk.AccAddress	`json:"acc_address"`
	PubKey     crypto.PubKey	`json:"pub_key"`
}

type MsgSwitchOff struct {
	AccAddress sdk.AccAddress `json:"acc_address"`
}

type MsgUnjail struct {
	AccAddress sdk.AccAddress `json:"acc_address"`
}

func NewMsgSwitchOn(accAddr sdk.AccAddress, pubKey crypto.PubKey) MsgSwitchOn {
	return MsgSwitchOn{
		AccAddress: accAddr,
		PubKey:     pubKey,
	}
}

func NewMsgSwitchOff(accAddr sdk.AccAddress) MsgSwitchOff {
	return MsgSwitchOff{
		AccAddress: accAddr,
	}
}

func NewMsgUnjail(accAddr sdk.AccAddress) MsgUnjail {
	return MsgUnjail{
		AccAddress: accAddr,
	}
}

const (
	SwitchOnConst  = "SwitchOn"
	SwitchOffConst = "SwitchOff"
	UnjailConst    = "Unjail"
)
// --- MsgSwitchOn implementation ---
// nolint
func (msg MsgSwitchOn) Route() string { return RouterKey }
func (msg MsgSwitchOn) Type() string  { return SwitchOnConst }
func (msg MsgSwitchOn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.AccAddress}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgSwitchOn) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgSwitchOn) ValidateBasic() error {
	if msg.AccAddress.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing account address")
	}
	if msg.PubKey == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "missing node public key")
	}
	return nil
}

// --- MsgSwitchOff implementation ---
func (msg MsgSwitchOff) Route() string { return RouterKey }
func (msg MsgSwitchOff) Type() string  { return SwitchOffConst }
func (msg MsgSwitchOff) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.AccAddress}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgSwitchOff) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgSwitchOff) ValidateBasic() error {
	if msg.AccAddress.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing account address")
	}
	return nil
}

// --- MsgUnjail implementation ---
func (msg MsgUnjail) Route() string { return RouterKey }
func (msg MsgUnjail) Type() string  { return UnjailConst }
func (msg MsgUnjail) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.AccAddress}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgUnjail) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgUnjail) ValidateBasic() error {
	if msg.AccAddress.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing account address")
	}
	return nil
}