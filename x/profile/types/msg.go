package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = MsgSetProfile{}

type MsgSetProfile struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Profile Profile        `json:"profile" yaml:"profile"`
}

func NewMsgSetProfile(addr sdk.AccAddress, profile Profile) MsgSetProfile {
	return MsgSetProfile{
		Address: addr,
		Profile: profile,
	}
}

// Route should return the name of the module
func (msg MsgSetProfile) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetProfile) Type() string { return "set_profile" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetProfile) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Address.String())
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetProfile) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetProfile) GetSigners() []sdk.AccAddress {
	//return []sdk.AccAddress{}/
	return []sdk.AccAddress{msg.Address}
}

type MsgSetNickname struct {
	Address  sdk.AccAddress `json:"address" yaml:"address"`
	Nickname string         `json:"nickname" yaml:"nickname"`
}

func NewMsgSetNickname(addr sdk.AccAddress, nickname string) MsgSetNickname {
	return MsgSetNickname{
		Address:  addr,
		Nickname: nickname,
	}
}

// Route should return the name of the module
func (msg MsgSetNickname) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetNickname) Type() string { return "set_nickname" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetNickname) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Address.String())
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetNickname) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetNickname) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

type MsgSetCardNumber struct {
	Address    sdk.AccAddress `json:"address" yaml:"address"`
	CardNumber uint64         `json:"card_number" yaml:"card_number"`
}

func NewMsgSetCardNumber(addr sdk.AccAddress, cardNumber uint64) MsgSetCardNumber {
	return MsgSetCardNumber{
		Address:    addr,
		CardNumber: cardNumber,
	}
}

// Route should return the name of the module
func (msg MsgSetCardNumber) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetCardNumber) Type() string { return "set_card_number" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetCardNumber) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Address.String())
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetCardNumber) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetCardNumber) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

// Create account with filled profiles
type MsgCreateAccountWithProfile struct {
	Address         sdk.AccAddress `json:"address" yaml:"address"`
	NewAccount      sdk.AccAddress `json:"new_account" yaml:"new_account"`
	ReferralAddress sdk.AccAddress `json:"referral" yaml:"referral"`
	Profile         Profile        `json:"profile" yaml:"profile"`
}

func NewMsgCreateAccountWithProfile(addr sdk.AccAddress, newAccount sdk.AccAddress, referralAddress sdk.AccAddress, profile Profile) MsgCreateAccountWithProfile {
	return MsgCreateAccountWithProfile{
		Address:         addr,
		NewAccount:      newAccount,
		ReferralAddress: referralAddress,
		Profile:         profile,
	}
}

// Route should return the name of the module
func (msg MsgCreateAccountWithProfile) Route() string { return RouterKey }

// Type should return the action
func (msg MsgCreateAccountWithProfile) Type() string { return "new_account_with_profile" }

// ValidateBasic runs stateless checks on the message
func (msg MsgCreateAccountWithProfile) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Address.String())
	}

	if msg.NewAccount.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.NewAccount.String())
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateAccountWithProfile) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgCreateAccountWithProfile) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

type MsgCreateAccount struct {
	Address         sdk.AccAddress `json:"address" yaml:"address"`
	NewAccount      sdk.AccAddress `json:"new_account" yaml:"new_account"`
	ReferralAddress sdk.AccAddress `json:"referral" yaml:"referral"`
}

func NewMsgCreateAccount(addr sdk.AccAddress, newAccount sdk.AccAddress, referralAddress sdk.AccAddress) MsgCreateAccount {
	return MsgCreateAccount{
		Address:         addr,
		NewAccount:      newAccount,
		ReferralAddress: referralAddress,
	}
}

// Route should return the name of the module
func (msg MsgCreateAccount) Route() string { return RouterKey }

// Type should return the action
func (msg MsgCreateAccount) Type() string { return "new_account" }

// ValidateBasic runs stateless checks on the message
func (msg MsgCreateAccount) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Address.String())
	}

	if msg.NewAccount.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.NewAccount.String())
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgCreateAccount) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}
