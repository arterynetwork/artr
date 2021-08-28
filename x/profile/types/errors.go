package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrNicknamePrefix       = sdkerrors.Register(ModuleName, 1, "nickname cannot start with 'ARTR-' prefix")
	ErrNicknameAlreadyInUse = sdkerrors.Register(ModuleName, 2, "nickname is already in use")
	ErrNotFound             = sdkerrors.Register(ModuleName, 3, "profile not found")
	ErrAccountAlreadyExists = sdkerrors.Register(ModuleName, 4, "account already exists")
	ErrNicknameTooShort     = sdkerrors.Register(ModuleName, 5, "nickname is too short")
	ErrNicknameInvalidChars = sdkerrors.Register(ModuleName, 6, "nickname contains invalid characters")
	ErrUnauthorized         = sdkerrors.Register(ModuleName, 7, "sender is out of whitelist")
)
