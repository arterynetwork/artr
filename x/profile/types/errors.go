package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrNicknamePrefix       = sdkerrors.Register(ModuleName, 1, "nickname cannot start with 'ARTR-' prefix")
	ErrNicknameAlreadyInUse = sdkerrors.Register(ModuleName, 2, "nickname is already in use")
)
