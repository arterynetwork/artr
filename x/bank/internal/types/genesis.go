package types

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	SendEnabled    bool  `json:"send_enabled" yaml:"send_enabled"`
	MinSend        int64 `json:"min_send" yaml:"min_send"`
	DustDelegation int64 `json:"dust_delegation" yaml:"dust_delegation"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(sendEnabled bool, minSend int64, dustDelegation int64) GenesisState {
	return GenesisState{SendEnabled: sendEnabled, MinSend: minSend, DustDelegation: dustDelegation}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState { return NewGenesisState(true, 1000, 0) }

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error { return nil }
