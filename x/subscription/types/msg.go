package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// verify interface at compile time
var _ sdk.Msg = &MsgPaySubscription{}

type MsgPaySubscription struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	StorageAmount int64          `json:"storage_amount" yaml:"storage_amount"`
}

// NewMsgPaySubscription creates a new Msg<Action> instance
func NewMsgPaySubscription(addr sdk.AccAddress, storageAmount int64) MsgPaySubscription {
	return MsgPaySubscription{
		Address:       addr,
		StorageAmount: storageAmount,
	}
}

const PaySubscriptionConst = "pay_subscription"

// nolint
func (msg MsgPaySubscription) Route() string { return RouterKey }
func (msg MsgPaySubscription) Type() string  { return PaySubscriptionConst }
func (msg MsgPaySubscription) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Address)}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgPaySubscription) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgPaySubscription) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing subscriber address")
	}
	return nil
}

type MsgPayVPN struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Amount  int64          `json:"amount" yaml:"amount"`
}

// NewMsgPayVPN creates a new Msg<Action> instance
func NewMsgPayVPN(addr sdk.AccAddress, amount int64) MsgPayVPN {
	return MsgPayVPN{
		Address: addr,
		Amount:  amount,
	}
}

const PayVPNConst = "pay_vpn"

// nolint
func (msg MsgPayVPN) Route() string { return RouterKey }
func (msg MsgPayVPN) Type() string  { return PayVPNConst }
func (msg MsgPayVPN) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Address)}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgPayVPN) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgPayVPN) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing subscriber address")
	}
	return nil
}

type MsgPayStorage struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Amount  int64          `json:"amount" yaml:"amount"`
}

// NewMsgPayStorage creates a new Msg<Action> instance
func NewMsgPayStorage(addr sdk.AccAddress, amount int64) MsgPayStorage {
	return MsgPayStorage{
		Address: addr,
		Amount:  amount,
	}
}

const PayStorageConst = "pay_storage"

// nolint
func (msg MsgPayStorage) Route() string { return RouterKey }
func (msg MsgPayStorage) Type() string  { return PayStorageConst }
func (msg MsgPayStorage) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Address)}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgPayStorage) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgPayStorage) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing subscriber address")
	}
	return nil
}

type MsgSetTokenRate struct {
	Sender sdk.AccAddress `json:"sender"`
	Value  uint32         `json:"value"`
}

func NewMsgSetTokenRate(sender sdk.AccAddress, value uint32) MsgSetTokenRate {
	return MsgSetTokenRate{
		Sender: sender,
		Value:  value,
	}
}

const SetTokenCourseRate = "set_token_rate"

// nolint
func (msg MsgSetTokenRate) Route() string { return RouterKey }
func (msg MsgSetTokenRate) Type() string  { return SetTokenCourseRate }
func (msg MsgSetTokenRate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgSetTokenRate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgSetTokenRate) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing subscriber address")
	}
	if msg.Value <= 0 { return fmt.Errorf("token exchange rate must be positive") }
	return nil
}
