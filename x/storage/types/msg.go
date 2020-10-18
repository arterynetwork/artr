package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// verify interface at compile time
var _ sdk.Msg = &MsgSetStorageData{}

// Msg<Action> - struct for unjailing jailed validator
type MsgSetStorageData struct {
	Address sdk.AccAddress `json:"address" yaml:"address"` // address of the validator operator
	Size    int64          `json:"size" yaml:"size"`
	Data    string         `json:"data" yaml:"data"`
}

// NewMsg<Action> creates a new Msg<Action> instance
func NewMsgSetStorageData(addr sdk.AccAddress, size int64, data string) MsgSetStorageData {
	return MsgSetStorageData{
		Address: addr,
		Size:    size,
		Data:    data,
	}
}

const SetStorageDataConst = "set_storage_data"

// nolint
func (msg MsgSetStorageData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetStorageData) Type() string { return SetStorageDataConst }

func (msg MsgSetStorageData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgSetStorageData) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgSetStorageData) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing address")
	}
	return nil
}
