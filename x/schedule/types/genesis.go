package types

import (
	"github.com/pkg/errors"
)

//
//// NewGenesisState creates a new GenesisState object
//func NewGenesisState(tasks []GenesisSchedule) GenesisState {
//	return GenesisState{
//		Tasks:  tasks,
//	}
//}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// ValidateGenesis validates the schedule genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return errors.Wrap(err, "invalid params")
	}
	for i, t := range data.Tasks {
		if t.HandlerName == "" {
			return errors.Errorf("empty handler at %s (#%d)", t.Time, i)
		}
	}
	return nil
}
