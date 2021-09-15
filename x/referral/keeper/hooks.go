package keeper

import (
	"github.com/arterynetwork/artr/x/referral/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	StatusUpdatedCallback = "status-updated"
	StakeChangedCallback  = "stake-changed"
	BanishedCallback      = "banished"

	StatusDowngradeHookName   = "referral/downgrade"
	CompressionHookName       = "referral/compression"
	TransitionTimeoutHookName = "referral/transition-timeout"
	BanishHookName            = "referral/banish"
)

func (k *Keeper) AddHook(eventName string, callback func(ctx sdk.Context, acc sdk.AccAddress) error) {
	lst, found := k.eventHooks[eventName]
	if !found {
		lst = make([]func(ctx sdk.Context, acc sdk.AccAddress) error, 0, 1)
	}
	lst = append(lst, callback)
	k.eventHooks[eventName] = lst
}

func (k Keeper) PerformDowngrade(ctx sdk.Context, data []byte) {
	if err := k.performDowngrade(ctx, sdk.AccAddress(data)); err != nil {
		panic(err)
	}
}

func (k Keeper) PerformCompression(ctx sdk.Context, data []byte) {
	if err := k.performCompression(ctx, sdk.AccAddress(data)); err != nil {
		panic(err)
	}
}

func (k Keeper) PerformTransitionTimeout(ctx sdk.Context, data []byte) {
	if err := k.CancelTransition(ctx, data, true); err != nil {
		panic(err)
	}
}

func (k Keeper) PerformBanish(ctx sdk.Context, data []byte) {
	if err := k.Banish(ctx, sdk.AccAddress(data)); err != nil {
		panic(err)
	}
}

func (k Keeper) callback(eventName string, ctx sdk.Context, acc sdk.AccAddress) error {
	lst, found := k.eventHooks[eventName]
	if !found {
		return nil
	}
	for _, hook := range lst {
		if err := hook(ctx, acc); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) performDowngrade(ctx sdk.Context, acc sdk.AccAddress) error {
	bu := newBunchUpdater(k, ctx)
	err := bu.update(acc, true, func(value *types.R) {
		if value.StatusDowngradeAt != ctx.BlockHeight() { // the user fixed things up
			return
		}
		bu.ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeStatusUpdated,
				sdk.NewAttribute(types.AttributeKeyAddress, acc.String()),
				sdk.NewAttribute(types.AttributeKeyStatusBefore, value.Status.String()),
				sdk.NewAttribute(types.AttributeKeyStatusAfter, (value.Status-1).String()),
			),
		)
		k.setStatus(ctx, value, value.Status-1, acc)
		value.StatusDowngradeAt = -1
	})
	if err != nil {
		return err
	}
	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}

func (k Keeper) performCompression(ctx sdk.Context, acc sdk.AccAddress) error {
	record, err := k.Get(ctx, acc)
	if err != nil {
		return err
	}
	if record.CompressionAt != ctx.BlockHeight() {
		return nil
	}

	return k.Compress(ctx, acc)
}
