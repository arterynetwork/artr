package types

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, earners []Earner) *GenesisState {
	return &GenesisState{
		Params:  params,
		Earners: earners,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// ValidateGenesis validates the earning genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return errors.Wrap(err, "invalid params")
	}
	for i, earner := range data.Earners {
		if _, err := sdk.AccAddressFromBech32(earner.Account); err != nil {
			return errors.Errorf("invalid earner #%d: invalid account", i)
		}
	}
	return nil
}
