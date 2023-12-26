package keeper

import (
	"sort"
	"time"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/delegating/types"
)

func (k Keeper) MustPerformRevoking(ctx sdk.Context, payload []byte, _ time.Time) {
	if err := k.performRevoking(ctx, payload); err != nil {
		k.Logger(ctx).Error("cannot perform revoke", "account", sdk.AccAddress(payload).String(), "error", err)
		panic(err)
	}
}

func (k Keeper) performRevoking(ctx sdk.Context, acc sdk.AccAddress) error {
	var (
		mainStore = ctx.KVStore(k.mainStoreKey)
		byteKey   = []byte(acc)

		byteItem []byte
		item     types.Record
		err      error
	)

	if !mainStore.Has(byteKey) {
		return nil
	}
	byteItem = mainStore.Get(byteKey)
	k.cdc.MustUnmarshalBinaryBare(byteItem, &item)
	sort.Slice(item.Requests, func(i, j int) bool {
		return item.Requests[i].Time.Before(item.Requests[j].Time)
	})

	n := 0
	for _, req := range item.Requests {
		if req.Time.After(ctx.BlockTime()) {
			break
		}

		if err = k.undelegate(ctx, acc, req.Amount); err != nil {
			return err
		}
		n += 1
	}
	if n == 0 {
		return nil
	}
	item.Requests = item.Requests[n:]
	if item.IsEmpty() {
		mainStore.Delete(byteKey)
	} else {
		bz := k.cdc.MustMarshalBinaryBare(&item)
		mainStore.Set(byteKey, bz)
	}
	return nil
}

func (k Keeper) MustPerformAccrue(ctx sdk.Context, payload []byte, time time.Time) {
	var (
		acc   sdk.AccAddress = payload
		store                = ctx.KVStore(k.mainStoreKey)
		data  types.Record
	)

	k.cdc.MustUnmarshalBinaryBare(store.Get(acc), &data)
	if data.NextAccrue == nil {
		panic(errors.New("accrue cancelled"))
	} else if *data.NextAccrue != time {
		panic(errors.Errorf("accrue rescheduled (%s ≠ %s)", data.NextAccrue, time))
	}

	delegated, _ := k.getDelegated(ctx, acc)
	isActiveProfile := k.profileKeeper.GetProfile(ctx, acc).IsActive(ctx)
	isActiveValidator, err := k.nodingKeeper.IsActiveValidator(ctx, acc)
	if err != nil {
		panic(err)
	}
	isActiveVpn, isActiveStorage, err := k.earningKeeper.IsActiveEarner(ctx, acc)
	if err != nil {
		panic(err)
	}
	bonusFlags := getBitmap(isActiveValidator, isActiveProfile, isActiveVpn, isActiveStorage)
	percent := k.percent(ctx, delegated, isActiveProfile, isActiveValidator, isActiveVpn, isActiveStorage)
	if percent.IsZero() {
		data.NextAccrue = nil
	} else {
		interest := percent.MulInt64(delegated.Int64()).Int64()
		interestToValidator := delegated.Int64()
		if data.MissedPart != nil {
			interest -= data.MissedPart.MulInt64(interest).Int64()
			interestToValidator -= data.MissedPart.MulInt64(interestToValidator).Int64()
			data.MissedPart = nil
		}
		k.accrue(ctx, acc, sdk.NewInt(interest), bonusFlags)
		k.accrueToValidator(ctx, acc, sdk.NewInt(interestToValidator))
		*data.NextAccrue = time.Add(k.scheduleKeeper.OneDay(ctx))
		k.scheduleKeeper.ScheduleTask(ctx, *data.NextAccrue, types.AccrueHookName, acc)
	}
	store.Set(acc, k.cdc.MustMarshalBinaryBare(&data))
}

func (k Keeper) OnBanished(ctx sdk.Context, acc sdk.AccAddress) error {
	d, _ := k.getDelegated(ctx, acc)
	if !d.IsZero() {
		if err := k.Revoke(ctx, acc, d); err != nil {
			return errors.Wrap(err, "cannot revoke delegation")
		}
	}
	return nil
}
