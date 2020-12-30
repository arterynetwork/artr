package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type GenesisProfile struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Profile Profile        `json:"profile" yaml:"profile"`
}

// GenesisState - all profile state that must be provided at genesis
type GenesisState struct {
	ProfileRecords []GenesisProfile `json:"profiles" yaml:"profiles"`
	Params         Params           `json:"params" yaml:"params"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, profiles []GenesisProfile) GenesisState {
	return GenesisState{
		ProfileRecords: profiles,
		Params:         params,
	}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the profile genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return nil
}
