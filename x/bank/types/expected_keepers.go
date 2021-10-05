package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) auth.AccountI

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.AccountI
	GetAllAccounts(ctx sdk.Context) []auth.AccountI
	SetAccount(ctx sdk.Context, acc auth.AccountI)

	IterateAccounts(ctx sdk.Context, process func(auth.AccountI) bool)

	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) auth.ModuleAccountI
}
