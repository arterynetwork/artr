package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type GenesisActivityInfo struct {
	Address      sdk.AccAddress `json:"address" yaml:"address"`
	ActivityInfo ActivityInfo   `json:"info" yaml:"info"`
}

// GenesisState - all subscription state that must be provided at genesis
type GenesisState struct {
	Params   Params                `json:"params" yaml:"params"`
	Activity []GenesisActivityInfo `json:"activity" yaml:"activity"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, activity []GenesisActivityInfo) GenesisState {
	return GenesisState{
		Params:   params,
		Activity: activity,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:   DefaultParams(),
		Activity: []GenesisActivityInfo{},
	}
}

// ValidateGenesis validates the subscription genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil { return err }
	for _, record := range data.Activity {
		if record.Address.Empty() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
		}
	}

	return nil
}
