package types

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, accounts []Account) GenesisState {
	return GenesisState{
		Params:   params,
		Accounts: accounts,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: *DefaultParams(),
	}
}

// ValidateGenesis validates the delegating genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return errors.Wrap(err, "invalid params")
	}
	for i, acc := range data.Accounts {
		if _, err := sdk.AccAddressFromBech32(acc.Address); err != nil {
			return errors.Wrapf(err, "invalid account #%d (%s)", i, acc.Address)
		}
		for j, revoke := range acc.Requests {
			if !revoke.Amount.IsPositive() {
				return errors.Errorf("invalid revoke #%d.%d (%s %s): amount is non-positive", i, j, acc.Address, revoke.Time.String())
			}
		}
	}
	return nil
}
