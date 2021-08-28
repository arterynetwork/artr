package keeper

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/delegating/types"
)

func (k Keeper) InitAccounts(ctx sdk.Context, accounts []types.Account) {
	store := ctx.KVStore(k.mainStoreKey)
	for _, account := range accounts {
		byteKey, err := sdk.AccAddressFromBech32(account.Address)
		if err != nil {
			k.Logger(ctx).Error("Invalid account address", "req", account, "err", err)
			panic(errors.Wrap(err, "cannot parse account address"))
		}

		item := types.Record{
			NextAccrue: account.NextAccrue,
			Requests:   account.Requests,
		}
		bz := k.cdc.MustMarshalBinaryBare(&item)
		store.Set(byteKey, bz)
	}
}

func (k Keeper) ExportAccounts(ctx sdk.Context) []types.Account {
	var result []types.Account
	store := ctx.KVStore(k.mainStoreKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key())
		var r types.Record
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &r)
		result = append(result, types.Account{
			Address:    acc.String(),
			NextAccrue: r.NextAccrue,
			Requests:   r.Requests,
		})
	}
	return result
}
