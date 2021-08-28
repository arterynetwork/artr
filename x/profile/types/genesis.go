package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

func (gp GenesisProfile) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(gp.Address)
	if err != nil {
		panic(errors.Wrapf(err, "invalid address %+v", gp))
	}
	return addr
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, profiles []GenesisProfile) *GenesisState {
	return &GenesisState{
		Profiles: profiles,
		Params:   params,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: *DefaultParams(),
	}
}

// ValidateGenesis validates the profile genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return nil
}
