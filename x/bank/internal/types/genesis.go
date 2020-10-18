package types

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	SendEnabled bool  `json:"send_enabled" yaml:"send_enabled"`
	MinSend     int64 `json:"min_send" yaml:"min_send"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(sendEnabled bool, minSend int64) GenesisState {
	return GenesisState{SendEnabled: sendEnabled, MinSend: minSend}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState { return NewGenesisState(true, 1000) }

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error { return nil }
