package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrDataToLong           = sdkerrors.Register(ModuleName, 1, "Directory data to long")
	ErrLimitSmallerThenData = sdkerrors.Register(ModuleName, 2, "Size limit smaller then data length")
	ErrLimitToLow           = sdkerrors.Register(ModuleName, 3, "Size limit smaller than minimal packet size")
)
