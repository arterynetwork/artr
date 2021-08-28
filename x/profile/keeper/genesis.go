package keeper

import (
	"strings"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/profile/types"
	ref "github.com/arterynetwork/artr/x/referral/types"
)

func (k Keeper) ExportProfileRecords(ctx sdk.Context) []types.GenesisProfile {
	var result []types.GenesisProfile
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key())
		var value types.Profile
		if err := k.cdc.UnmarshalBinaryBare(it.Value(), &value); err != nil {
			panic(err)
		}
		value.CardNumber = 0
		result = append(result, types.GenesisProfile{
			Address: acc.String(),
			Profile: value,
		})
	}
	return result
}

func (k Keeper) ImportProfileRecords(ctx sdk.Context, data []types.GenesisProfile) {
	k.Logger(ctx).Info("... user profiles")
	for _, record := range data {
		addr := record.GetAddress()
		acc := k.accountKeeper.GetAccount(ctx, addr)
		record.Profile.CardNumber = k.CardNumberByAccountNumber(ctx, acc.GetAccountNumber())

		nickname := strings.TrimSpace(record.Profile.Nickname)
		if nickname != "" {
			k.setProfileAccountByNickname(ctx, nickname, addr)
		}
		k.setProfileAccountByCardNumber(ctx, record.Profile.CardNumber, addr)

		store := ctx.KVStore(k.storeKey)
		bz, err := k.cdc.MarshalBinaryBare(&record.Profile)
		if err != nil {
			panic(err)
		}
		store.Set(addr, bz)

		if acc == nil {
			acc = k.accountKeeper.NewAccountWithAddress(ctx, addr)
			k.accountKeeper.SetAccount(ctx, acc)
		}

		if active := record.Profile.IsActive(ctx); active {
			k.referralKeeper.MustSetActiveWithoutStatusUpdate(ctx, addr.String(), active)
		}

		if err := k.SetProfile(ctx, addr, record.Profile); err != nil {
			panic(errors.Wrapf(err, "invalid profile %s", record.Address))
		}
	}
	// Now, when all activity flags and referral counts are set, update all statuses at once
	k.Logger(ctx).Info("... referral statuses")
	k.referralKeeper.Iterate(ctx, func(_ string, _ *ref.Info) (changed, checkForStatusUpdate bool) {
		return false, true
	})
}