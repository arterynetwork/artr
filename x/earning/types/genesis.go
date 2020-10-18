package types

import "fmt"

// GenesisState - all earning state that must be provided at genesis
type GenesisState struct {
	Params  Params      `json:"params"`
	State   StateParams `json:"state,omitempty"`
	Earners []Earner    `json:"earners,omitempty"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, state StateParams, earners []Earner) GenesisState {
	return GenesisState{
		Params:  params,
		State:   state,
		Earners: earners[:],
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// ValidateGenesis validates the earning genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil { return err }
	if err := data.State.Validate(); err != nil { return err }
	for i, earner := range data.Earners {
		if earner.Account.Empty() { return fmt.Errorf("account is empty (#%d", i) }
		if earner.Vpn < 0 { return fmt.Errorf("vpn points must be non-negative") }
		if earner.Storage < 0 { return fmt.Errorf("storage points must be non-negative") }
	}
	return nil
}
