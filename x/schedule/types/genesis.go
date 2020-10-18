package types

import "fmt"

// GenesisState - all schedule state that must be provided at genesis
type GenesisState struct {
	Tasks []GenesisSchedule `json:"tasks"`
}

type GenesisSchedule struct {
	Schedule
	Height uint64 `json:"height"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(tasks []GenesisSchedule) GenesisState {
	return GenesisState{
		Tasks: tasks,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// ValidateGenesis validates the schedule genesis parameters
func ValidateGenesis(data GenesisState) error {
	keys := make(map[uint64]bool, len(data.Tasks))
	for _, t := range data.Tasks {
		if keys[t.Height] { return fmt.Errorf("duplicating key: %d", t.Height)}
		keys[t.Height] = true
	}
	return nil
}
