package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNothingDelegated = sdkerrors.Register(ModuleName, 1, "nothing's delegated")
	ErrLessThanMinimum  = sdkerrors.Register(ModuleName, 2, "delegation is lass than minimum")
)
