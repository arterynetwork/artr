package keeper

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/bank/types"
)

var _ ViewKeeper = (*BaseViewKeeper)(nil)

type ViewKeeper interface {
	ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool

	GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetAccountsBalances(ctx sdk.Context) []types.Balance
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coins) (stop bool))
}

// BaseViewKeeper implements a read only keeper implementation of ViewKeeper.
type BaseViewKeeper struct {
	cdc      codec.BinaryMarshaler
	storeKey sdk.StoreKey
	ak       types.AccountKeeper
}

// NewBaseViewKeeper returns a new BaseViewKeeper.
func NewBaseViewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, ak types.AccountKeeper) BaseViewKeeper {
	return BaseViewKeeper{
		cdc:      cdc,
		storeKey: storeKey,
		ak:       ak,
	}
}

// Logger returns a module-specific logger.
func (keeper BaseViewKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HasBalance returns whether or not an account has at least amt balance.
func (k BaseViewKeeper) HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool {
	return k.GetBalance(ctx, addr).IsAllGTE([]sdk.Coin{amt})
}

// GetBalance returns all the account balances for the given account address.
func (k BaseViewKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	key := make([]byte, len(types.BalancesPrefix)+len(addr.Bytes()))
	copy(key, types.BalancesPrefix)
	copy(key[len(types.BalancesPrefix):], addr.Bytes())

	bz := store.Get(key)
	if bz == nil {
		return sdk.Coins{}
	}

	var balance types.Balance
	if err := k.cdc.UnmarshalBinaryBare(bz, &balance); err != nil { panic(errors.Wrap(err, "cannot unmarshal value")) }
	return sdk.NewCoins(balance.Coins...)
}

// GetAccountsBalances returns all the accounts balances from the store.
func (k BaseViewKeeper) GetAccountsBalances(ctx sdk.Context) []types.Balance {
	balances := make([]types.Balance, 0)

	k.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coins) bool {
		balances = append(balances, types.Balance{
			Address: addr.String(),
			Coins:   balance,
		})
		return false
	})

	return balances
}

// IterateAllBalances iterates over all the balances of all accounts and
// denominations that are provided to a callback. If true is returned from the
// callback, iteration is halted.
func (k BaseViewKeeper) IterateAllBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coins) bool) {
	store := ctx.KVStore(k.storeKey)
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)

	iterator := balancesStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address := types.AddressFromBalancesStore(iterator.Key())

		var balance types.Balance
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &balance)

		if cb(address, balance.Coins) {
			break
		}
	}
}

// LockedCoins returns all the coins that are not spendable (i.e. locked) for an
// account by address. For standard accounts, the result will always be no coins.
// For vesting accounts, LockedCoins is delegated to the concrete vesting account
// type.
func (k BaseViewKeeper) LockedCoins(_ sdk.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

// SpendableCoins returns the total balances of spendable coins for an account
// by address. If the account has no spendable coins, an empty Coins slice is
// returned.
func (k BaseViewKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.GetBalance(ctx, addr)
}

// ValidateBalance validates all balances for a given account address returning
// an error if any balance is invalid. It will check for vesting account types
// and validate the balances against the original vesting balances.
//
// CONTRACT: ValidateBalance should only be called upon genesis state. In the
// case of vesting accounts, balances may change in a valid manner that would
// otherwise yield an error from this call.
func (k BaseViewKeeper) ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error {
	acc := k.ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	balances := k.GetBalance(ctx, addr)
	if !balances.IsValid() {
		return fmt.Errorf("account balance of %s is invalid", balances)
	}

	return nil
}
