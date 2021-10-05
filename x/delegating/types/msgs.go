package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// verify interface at compile time
var _ sdk.Msg = &MsgDelegate{}

// NewMsgDelegate creates a new MsgDelegate instance
func NewMsgDelegate(acc sdk.AccAddress, ucoins sdk.Int) MsgDelegate {
	return MsgDelegate{
		Address:    acc.String(),
		MicroCoins: ucoins,
	}
}

const DelegateConst = "delegate"

// nolint
func (msg MsgDelegate) Route() string { return RouterKey }
func (msg MsgDelegate) Type() string  { return DelegateConst }
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{address}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgDelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid account address")
	}
	if !msg.MicroCoins.IsPositive() {
		return errors.New("amount must be positive")
	}
	return nil
}

func (msg MsgDelegate) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

// verify interface at compile time
var _ sdk.Msg = &MsgRevoke{}

// NewMsgDelegate creates a new MsgDelegate instance
func NewMsgRevoke(acc sdk.AccAddress, ucoins sdk.Int) MsgRevoke {
	return MsgRevoke{
		Address:    acc.String(),
		MicroCoins: ucoins,
	}
}

const RevokeConst = "revoke"

// nolint
func (msg MsgRevoke) Route() string { return RouterKey }
func (msg MsgRevoke) Type() string  { return RevokeConst }
func (msg MsgRevoke) GetSigners() []sdk.AccAddress {
	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{address}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgRevoke) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgRevoke) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid account address")
	}
	if !msg.MicroCoins.IsPositive() {
		return errors.New("amount must be positive")
	}
	return nil
}

func (msg MsgRevoke) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}
