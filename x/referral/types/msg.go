package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = new(MsgRequestTransition)
	_ sdk.Msg = new(MsgResolveTransition)
)

const (
	RequestTransitionConst = "RequestTransition"
	ResolveTransitionConst = "ResolveTransition"
)

type MsgRequestTransition struct {
	Subject     sdk.AccAddress `json:"subject"`
	Destination sdk.AccAddress `json:"destination"`
}

func NewMsgRequestTransition(subject, destination sdk.AccAddress) MsgRequestTransition {
	return MsgRequestTransition{
		Subject:     subject,
		Destination: destination,
	}
}

type MsgResolveTransition struct {
	Sender   sdk.AccAddress `json:"sender"`
	Subject  sdk.AccAddress `json:"subject"`
	Approved bool           `json:"approved"`
}

func NewMsgResolveTransition(sender, subject sdk.AccAddress, approved bool) MsgResolveTransition {
	return MsgResolveTransition{
		Sender:   sender,
		Subject:  subject,
		Approved: approved,
	}
}

func (MsgRequestTransition) Route() string { return RouterKey }
func (MsgRequestTransition) Type() string  { return RequestTransitionConst }

func (msg MsgRequestTransition) ValidateBasic() error {
	if msg.Subject.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing subject address")
	}
	if msg.Destination.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing destination address")
	}
	return nil
}

func (msg MsgRequestTransition) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRequestTransition) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Subject}
}

func (MsgResolveTransition) Route() string { return RouterKey }
func (MsgResolveTransition) Type() string  { return ResolveTransitionConst }

func (msg MsgResolveTransition) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if msg.Subject.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing subject address")
	}
	return nil
}

func (msg MsgResolveTransition) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgResolveTransition) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
