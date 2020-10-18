package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNotQualified      = sdkerrors.Register(ModuleName, 1, "account is not qualified for noding")
	ErrPubkeyBusy        = sdkerrors.Register(ModuleName, 2, "node with this public key is already validator")
	ErrNotFound          = sdkerrors.Register(ModuleName, 3, "cannot find account data")
	ErrNotJailed         = sdkerrors.Register(ModuleName, 4, "validator is not jailed")
	ErrJailPeriodNotOver = sdkerrors.Register(ModuleName, 5, "jail period is not finished yet")
	ErrBannedForLifetime = sdkerrors.Register(ModuleName, 6, "validator is banned for a lifetime")
	ErrAlreadyOn         = sdkerrors.Register(ModuleName, 7, "noding is already on")
)
