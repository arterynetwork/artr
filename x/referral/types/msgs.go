package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

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

func NewMsgRequestTransition(subject, destination string) *MsgRequestTransition {
	return &MsgRequestTransition{
		Subject:     subject,
		Destination: destination,
	}
}

func NewMsgResolveTransition(sender, subject string, approved bool) *MsgResolveTransition {
	return &MsgResolveTransition{
		Signer:  sender,
		Subject: subject,
		Decline: !approved,
	}
}

func (msg MsgRequestTransition) GetSubject() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Subject)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgRequestTransition) GetDestination() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Destination)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgRequestTransition) Route() string { return RouterKey }
func (MsgRequestTransition) Type() string  { return RequestTransitionConst }

func (msg MsgRequestTransition) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Subject); err != nil {
		return errors.Wrap(err, "invalid subject address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Destination); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid destination address")
	}
	return nil
}

func (msg MsgRequestTransition) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}

func (msg MsgRequestTransition) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Subject)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgResolveTransition) GetSubject() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Subject)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgResolveTransition) GetSigner() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgResolveTransition) GetDeclined() bool {
	return msg.Decline
}

func (msg MsgResolveTransition) GetApproved() bool {
	return !msg.Decline
}

func (MsgResolveTransition) Route() string { return RouterKey }
func (MsgResolveTransition) Type() string  { return ResolveTransitionConst }

func (msg MsgResolveTransition) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return errors.Wrap(err, "invalid signer address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Subject); err != nil {
		return errors.Wrap(err, "invalid subject address")
	}
	return nil
}

func (msg MsgResolveTransition) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}

func (msg MsgResolveTransition) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSigner()}
}
