package types

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCreateAccount{}
	_ sdk.Msg = &MsgUpdateProfile{}
	_ sdk.Msg = &MsgSetStorageCurrent{}
	_ sdk.Msg = &MsgSetVpnCurrent{}
	_ sdk.Msg = &MsgPayTariff{}
	_ sdk.Msg = &MsgBuyStorage{}
	_ sdk.Msg = &MsgGiveStorageUp{}
	_ sdk.Msg = &MsgBuyVpn{}
	_ sdk.Msg = &MsgSetRate{}
	_ sdk.Msg = &MsgBuyImExtraStorage{}
	_ sdk.Msg = &MsgGiveUpImExtra{}
	_ sdk.Msg = &MsgProlongImExtra{}
)

const (
	CreateAccountConst     = "new_account"
	UpdateProfileConst     = "update_profile"
	SetStorageCurrentConst = "set_storage"
	SetVpnCurrentConst     = "set_vpn"
	PayTariffConst         = "pay_tariff"
	BuyStorageConst        = "buy_storage"
	GiveStorageUpConst     = "give_storage_up"
	BuyVpnConst            = "buy_vpn"
	SetRateConst           = "set_rate"
	BuyImExtraStorageConst = "buy_im_extra"
	GiveUpImExtraConst     = "give_up_im_extra"
	ProlongImExtraConst    = "prolong_im_extra"

	ForbiddenNicknameCharacters = " */:'\"=[],."
)

func NewMsgCreateAccount(creator sdk.AccAddress, account sdk.AccAddress, referrer sdk.AccAddress) MsgCreateAccount {
	return MsgCreateAccount{
		Creator:  creator.String(),
		Address:  account.String(),
		Referrer: referrer.String(),
	}
}

// Route should return the name of the module
func (msg MsgCreateAccount) Route() string { return RouterKey }

// Type should return the action
func (msg MsgCreateAccount) Type() string { return CreateAccountConst }

// ValidateBasic runs stateless checks on the message
func (msg MsgCreateAccount) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errors.Wrap(err, "invalid creator")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Referrer); err != nil {
		return errors.Wrap(err, "invalid referrer")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateAccount) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// GetSigners defines whose signature is required
func (msg MsgCreateAccount) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetCreator()}
}

func (msg MsgCreateAccount) GetCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgCreateAccount) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgCreateAccount) GetReferrer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Referrer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgCreateAccount) WithProfile() bool {
	return msg.Profile != nil
}

// Route should return the name of the module
func (msg MsgUpdateProfile) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUpdateProfile) Type() string { return UpdateProfileConst }

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateProfile) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return sdkerrors.Wrap(err, "invalid address")
	}

	if len(msg.Updates) == 0 {
		return errors.New("no updates")
	}
	for i, upd := range msg.Updates {
		switch upd.Field {
		case MsgUpdateProfile_Update_FIELD_NICKNAME:
			val, ok := upd.Value.(*MsgUpdateProfile_Update_String_)
			if !ok {
				return errors.Errorf("wrong update #%d value type: %T (string expected)", i, upd.Value)
			}
			s := val.String_
			if len(s) != 0 {
				if strings.ContainsAny(s, ForbiddenNicknameCharacters) {
					return sdkerrors.Wrapf(ErrNicknameInvalidChars, "wrong update #%d value", i)
				}
				if len(s) < 3 {
					return sdkerrors.Wrapf(ErrNicknameTooShort, "wrong update #%d value", i)
				}
				if strings.HasPrefix(s, "ARTR-") {
					return sdkerrors.Wrapf(ErrNicknamePrefix, "wrong update #%d value", i)
				}
			}
		case
			MsgUpdateProfile_Update_FIELD_AUTO_PAY,
			MsgUpdateProfile_Update_FIELD_NODING,
			MsgUpdateProfile_Update_FIELD_STORAGE,
			MsgUpdateProfile_Update_FIELD_VALIDATOR,
			MsgUpdateProfile_Update_FIELD_VPN:

			_, ok := upd.Value.(*MsgUpdateProfile_Update_Bool)
			if !ok {
				return errors.Errorf("wrong update #%d value type: %T (bool expected)", i, upd.Value)
			}
		}
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgUpdateProfile) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// GetSigners defines whose signature is required
func (msg MsgUpdateProfile) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAddress()}
}

func (msg MsgUpdateProfile) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

// Route should return the name of the module
func (msg MsgSetStorageCurrent) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetStorageCurrent) Type() string { return SetStorageCurrentConst }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetStorageCurrent) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if msg.Value < 0 {
		return errors.New("value must be non-negative")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetStorageCurrent) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// GetSigners defines whose signature is required
func (msg MsgSetStorageCurrent) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSender()}
}

func (msg MsgSetStorageCurrent) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgSetStorageCurrent) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgSetVpnCurrent) Route() string { return RouterKey }

func (MsgSetVpnCurrent) Type() string { return SetVpnCurrentConst }

func (msg MsgSetVpnCurrent) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if msg.Value < 0 {
		return errors.New("value must be non-negative")
	}

	return nil
}

func (msg *MsgSetVpnCurrent) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgSetVpnCurrent) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSender()}
}

func (msg MsgSetVpnCurrent) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgSetVpnCurrent) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgPayTariff) Route() string { return RouterKey }

func (MsgPayTariff) Type() string { return PayTariffConst }

func (msg MsgPayTariff) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	return nil
}

func (msg *MsgPayTariff) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgPayTariff) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAddress()}
}

func (msg MsgPayTariff) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgBuyStorage) Route() string { return RouterKey }

func (MsgBuyStorage) Type() string { return BuyStorageConst }

func (msg MsgBuyStorage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if msg.ExtraStorage <= 0 {
		return errors.New("extra_storage must be positive")
	}
	return nil
}

func (msg *MsgBuyStorage) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgBuyStorage) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAddress()}
}

func (msg MsgBuyStorage) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgGiveStorageUp) Route() string { return RouterKey }

func (MsgGiveStorageUp) Type() string { return GiveStorageUpConst }

func (msg MsgGiveStorageUp) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if msg.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}

func (msg *MsgGiveStorageUp) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgGiveStorageUp) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAccAddress()}
}

func (msg MsgGiveStorageUp) GetAccAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (msg MsgBuyVpn) Route() string { return RouterKey }

func (msg MsgBuyVpn) Type() string { return BuyStorageConst }

func (msg MsgBuyVpn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if msg.ExtraTraffic <= 0 {
		return errors.New("extra_traffic must be positive")
	}
	return nil
}

func (msg *MsgBuyVpn) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgBuyVpn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAddress()}
}

func (msg MsgBuyVpn) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgSetRate) Route() string { return RouterKey }

func (MsgSetRate) Type() string { return SetRateConst }

func (msg MsgSetRate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender")
	}
	if !msg.Value.IsPositive() {
		return errors.New("value must be positive")
	}
	return nil
}

func (msg *MsgSetRate) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgSetRate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSender()}
}

func (msg MsgSetRate) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgBuyImExtraStorage) Route() string { return RouterKey }

func (MsgBuyImExtraStorage) Type() string { return BuyImExtraStorageConst }

func (msg MsgBuyImExtraStorage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if msg.ExtraStorage <= 0 {
		return errors.New("extra_storage must be positive")
	}
	return nil
}

func (msg *MsgBuyImExtraStorage) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgBuyImExtraStorage) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAddress()}
}

func (msg MsgBuyImExtraStorage) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgGiveUpImExtra) Route() string { return RouterKey }

func (MsgGiveUpImExtra) Type() string { return GiveUpImExtraConst }

func (msg MsgGiveUpImExtra) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	if msg.Amount < 0 {
		return errors.New("amount must be non-negative")
	}
	return nil
}

func (msg *MsgGiveUpImExtra) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgGiveUpImExtra) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAddress()}
}

func (msg MsgGiveUpImExtra) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgProlongImExtra) Route() string { return RouterKey }

func (MsgProlongImExtra) Type() string { return ProlongImExtraConst }

func (msg MsgProlongImExtra) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return errors.Wrap(err, "invalid address")
	}
	return nil
}

func (msg *MsgProlongImExtra) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgProlongImExtra) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAddress()}
}

func (msg MsgProlongImExtra) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return addr
}
