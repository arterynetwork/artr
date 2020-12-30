package types

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Refs struct {
	Referrer  sdk.AccAddress   `json:"referrer"`
	Referrals []sdk.AccAddress `json:"referrals"`
}

type GenesisCompression struct {
	Account sdk.AccAddress `json:"account"`
	Height  int64          `json:"height"`
}

func NewGenesisCompression(acc sdk.AccAddress, at int64) GenesisCompression {
	return GenesisCompression{
		Account: acc,
		Height:  at,
	}
}

type GenesisStatusDowngrade struct {
	Account sdk.AccAddress `json:"account"`
	Current uint8          `json:"current"`
	Height  int64          `json:"height"`
}

func NewGenesisStatusDowngrade(acc sdk.AccAddress, current Status, at int64) GenesisStatusDowngrade {
	return GenesisStatusDowngrade{
		Account: acc,
		Current: uint8(current),
		Height:  at,
	}
}

// Transition represents an account transition request. Destination is equal to Subject's R.Transition field.
type Transition struct {
	Subject     sdk.AccAddress `json:"subj"`
	Destination sdk.AccAddress `json:"dest"`
}

// NewTransition creates and fully initializes a new Transition instance.
func NewTransition(subject, destination sdk.AccAddress) Transition {
	return Transition{
		Subject:     subject,
		Destination: destination,
	}
}

// Validate performs very basic sanity check.
func (t Transition) Validate() error {
	if t.Subject.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "subject address is missing")
	}
	if t.Destination.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "destination address is missing")
	}
	return nil
}

// GenesisState - all referral state that must be provided at genesis
type GenesisState struct {
	Params           Params                   `json:"params"`
	TopLevelAccounts []sdk.AccAddress         `json:"top_level_accounts"`
	OtherAccounts    []Refs                   `json:"other_accounts"`
	Compression      []GenesisCompression     `json:"compression,omitempty"`
	Downgrade        []GenesisStatusDowngrade `json:"downgrade,omitempty"`
	Transitions      []Transition             `json:"transitions,omitempty"`
}


// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	topLevelAccounts []sdk.AccAddress,
	otherAccounts []Refs,
	compressions []GenesisCompression,
	downgrades []GenesisStatusDowngrade,
	transitions []Transition,
) GenesisState {
	return GenesisState{
		Params:           params,
		TopLevelAccounts: topLevelAccounts,
		OtherAccounts:    otherAccounts,
		Compression:      compressions,
		Downgrade:        downgrades,
		Transitions:      transitions,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the referral genesis parameters
func ValidateGenesis(data GenesisState) error {
	if data.TopLevelAccounts == nil {
		return errors.New("empty top level accounts set")
	}
	if err := data.Params.Validate(); err != nil {
		return err
	}
	for i, t := range data.Transitions {
		if err := t.Validate(); err != nil {
			return errors.Wrapf(err, "invalid transition #%d", i)
		}
	}
	return nil
}
