package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// verify interface at compile time
var _ MsgEarningCommandI = &MsgSet{}
var _ MsgEarningCommandI = &MsgSetMultiple{}

type MsgEarningCommandI interface {
	sdk.Msg
	GetSigner() sdk.AccAddress
}

func NewMsgSet(sender sdk.AccAddress, earner Earner) *MsgSet {
	return &MsgSet{
		Earner: earner,
		Signer: sender.String(),
	}
}

func NewMsgSetMultiple(sender sdk.AccAddress, earners []Earner) *MsgSetMultiple {
	return &MsgSetMultiple{
		Earners: earners,
		Signer:  sender.String(),
	}
}

const SetConst = "set"
const SetMultipleConst = "set-multiple"

func (msg MsgSet) GetSigner() sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return acc
}

func (msg MsgSet) Route() string {
	return RouterKey
}

func (msg MsgSet) Type() string {
	return SetConst
}

func (msg MsgSet) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSigner()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgSet) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgSet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return errors.Wrap(err, "invalid signer")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Earner.Account); err != nil {
		return errors.Wrapf(err, "invalid earner acc address")
	}
	return nil
}

func (msg MsgSetMultiple) GetSigner() sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return acc
}

func (msg MsgSetMultiple) Route() string {
	return RouterKey
}

func (msg MsgSetMultiple) Type() string {
	return SetMultipleConst
}

func (msg MsgSetMultiple) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSigner()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgSetMultiple) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgSetMultiple) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return errors.Wrap(err, "invalid signer")
	}
	if len(msg.Earners) == 0 {
		return errors.New("missing earners list")
	}
	for i, earner := range msg.Earners {
		if _, err := sdk.AccAddressFromBech32(earner.Account); err != nil {
			return errors.Wrapf(err, "invalid earner #%d acc address", i)
		}
	}
	return nil
}
