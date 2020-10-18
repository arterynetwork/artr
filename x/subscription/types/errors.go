package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// You can see how they are constructed below:
var (
	ErrInactiveSubscription = sdkerrors.Register(ModuleName, 1, "Subscription is inactive")
)
