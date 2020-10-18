package types

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Refs struct {
	Referrer  sdk.AccAddress `json:"referrer"`
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

// GenesisState - all referral state that must be provided at genesis
type GenesisState struct {
	Params           	Params           	     `json:"params"`
	TopLevelAccounts 	[]sdk.AccAddress 	     `json:"top_level_accounts"`
	OtherAccounts 		[]Refs 				     `json:"other_accounts"`
	Compression         []GenesisCompression     `json:"compression,omitempty"`
	Downgrade           []GenesisStatusDowngrade `json:"downgrade,omitempty"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	topLevelAccounts []sdk.AccAddress,
	otherAccounts    []Refs,
	compressions     []GenesisCompression,
	downgrades       []GenesisStatusDowngrade,
) GenesisState {
	return GenesisState{
		Params:           params,
		TopLevelAccounts: topLevelAccounts,
		OtherAccounts:    otherAccounts,
		Compression:      compressions,
		Downgrade:        downgrades,
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
	if err := data.Params.Validate(); err != nil { return err }
	return nil
}
