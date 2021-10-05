package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrParentNil          = sdkerrors.Register(ModuleName, 1, "parentAcc cannot be nil")
	ErrRegistrationClosed = sdkerrors.Register(ModuleName, 2, "referrer is inactive for too long")
	ErrNotFound           = sdkerrors.Register(ModuleName, 3, "account is out of the referral structure")
)
