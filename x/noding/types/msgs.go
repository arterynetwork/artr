package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// verify interface at compile time
var (
	_ sdk.Msg = &MsgOn{}
	_ sdk.Msg = &MsgOff{}
	_ sdk.Msg = &MsgUnjail{}
)

func (msg MsgOn) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m MsgOn) GetPubKey() crypto.PubKey {
	return sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, m.PubKey)
}

func (m MsgOff) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m MsgUnjail) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func NewMsgOn(accAddr sdk.AccAddress, pubKey crypto.PubKey) *MsgOn {
	return &MsgOn{
		Account: accAddr.String(),
		PubKey:  sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, pubKey),
	}
}

func NewMsgOff(accAddr sdk.AccAddress) *MsgOff {
	return &MsgOff{
		Account: accAddr.String(),
	}
}

func NewMsgUnjail(accAddr sdk.AccAddress) *MsgUnjail {
	return &MsgUnjail{
		Account: accAddr.String(),
	}
}

const (
	SwitchOnConst  = "SwitchOn"
	SwitchOffConst = "SwitchOff"
	UnjailConst    = "Unjail"
)

func (MsgOn) Route() string { return RouterKey }
func (MsgOn) Type() string  { return SwitchOnConst }
func (msg MsgOn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAccount()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgOn) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgOn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return errors.Wrap(err, "invalid account")
	}
	if _, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, msg.PubKey); err != nil {
		return errors.Wrap(err, "invalid pub_key")
	}
	return nil
}

func (MsgOff) Route() string { return RouterKey }
func (MsgOff) Type() string  { return SwitchOffConst }
func (msg MsgOff) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAccount()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgOff) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgOff) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return errors.Wrap(err, "invalid account")
	}
	return nil
}

func (MsgUnjail) Route() string { return RouterKey }
func (MsgUnjail) Type() string  { return UnjailConst }
func (msg MsgUnjail) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAccount()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgUnjail) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgUnjail) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return errors.Wrap(err, "invalid account")
	}
	return nil
}
