package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all storage state that must be provided at genesis
type GenesisState struct {
	Limits  []Volume `json:"limits"`
	Current []Volume `json:"current"`
	Data    []Data   `json:"data"`
}

type Volume struct {
	Account sdk.AccAddress `json:"account"`
	Volume  uint64          `json:"volume"`
}

type Data struct {
	Account sdk.AccAddress `json:"account"`
	Base64  string         `json:"base64"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(limits []Volume, current []Volume, data []Data) GenesisState {
	return GenesisState{
		Limits:  limits,
		Current: current,
		Data:    data,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// ValidateGenesis validates the storage genesis parameters
func ValidateGenesis(data GenesisState) error {
	for i, limit := range data.Limits {
		if limit.Account.Empty() {
			return fmt.Errorf("empty account address (#%d)", i)
		}
	}
	return nil
}
