package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrAlreadyListed = sdkerrors.Register(ModuleName, 1, "account is already in list")
	ErrTooLate       = sdkerrors.Register(ModuleName, 2, "too late, cannot schedule for the past")
	ErrLocked        = sdkerrors.Register(ModuleName, 3, "earner list is locked")
	ErrNotLocked     = sdkerrors.Register(ModuleName, 4, "earner list is not locked")
	ErrNoMoney       = sdkerrors.Register(ModuleName, 5, "there are no coins in VPN&storage module accounts")
)
