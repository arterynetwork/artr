package keeper

import (
	"fmt"

	"github.com/arterynetwork/artr/x/bank/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "nonnegative-outstanding",
		NonnegativeBalanceInvariant(k))
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(k ViewKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var count int

		k.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coins) bool {
			if balance.IsAnyNegative() {
				count++
				msg += fmt.Sprintf("\t%s has a negative balance of %s\n", addr, balance)
			}

			return false
		})
		broken := count != 0

		return sdk.FormatInvariant(types.ModuleName, "nonnegative-outstanding",
			fmt.Sprintf("amount of negative accounts found %d\n%s", count, msg)), broken
	}
}
