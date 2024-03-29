package keeper

import (
	"time"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral/types"
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

// TODO: refactor x/noding too
func (k *Keeper) AddHook(eventName string, callback func(ctx sdk.Context, acc sdk.AccAddress) error) {
	lst, found := k.eventHooks[eventName]
	if !found {
		lst = make([]func(ctx sdk.Context, acc string) error, 0, 1)
	}
	lst = append(lst, func(ctx sdk.Context, acc string) error {
		if acc, err := sdk.AccAddressFromBech32(acc); err != nil {
			return errors.Wrap(err, "invalid account address")
		} else {
			return callback(ctx, acc)
		}
	})
	k.eventHooks[eventName] = lst
}

func (k Keeper) PerformDowngrade(ctx sdk.Context, data []byte, _ time.Time) {
	if err := k.performDowngrade(ctx, string(data)); err != nil {
		panic(err)
	}
}

func (k Keeper) PerformCompression(ctx sdk.Context, data []byte, _ time.Time) {
	if err := k.performCompression(ctx, string(data)); err != nil {
		panic(err)
	}
}

func (k Keeper) PerformTransitionTimeout(ctx sdk.Context, data []byte, _ time.Time) {
	if err := k.CancelTransition(ctx, string(data), true); err != nil {
		panic(err)
	}
}

func (k Keeper) PerformBanish(ctx sdk.Context, data []byte, _ time.Time) {
	if err := k.Banish(ctx, string(data)); err != nil {
		panic(err)
	}
}

func (k Keeper) callback(eventName string, ctx sdk.Context, acc string) error {
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

func (k Keeper) performDowngrade(ctx sdk.Context, acc string) error {
	k.Logger(ctx).Debug("performDowngrade", "acc", acc)
	bu := newBunchUpdater(k, ctx)
	err := bu.update(acc, true, func(value *types.Info) error {
		if value.StatusDowngradeAt == nil || value.StatusDowngradeAt.After(ctx.BlockTime()) { // the user fixed things up
			return nil
		}
		nextStatus := value.Status - 1
		if nextStatus == types.HeroDeprecatedStatus {
			nextStatus = nextStatus - 1
		}
		util.EmitEvent(bu.ctx,
			&types.EventStatusUpdated{
				Address: acc,
				Before:  value.Status,
				After:   nextStatus,
			},
		)
		k.setStatus(ctx, value, nextStatus, acc)
		value.StatusDowngradeAt = nil
		return nil
	})
	if err != nil {
		return err
	}
	bu.addCallback(StatusUpdatedCallback, acc)
	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}

func (k Keeper) performCompression(ctx sdk.Context, acc string) error {
	record, err := k.Get(ctx, acc)
	if err != nil {
		return err
	}
	if record.CompressionAt == nil || record.CompressionAt.After(ctx.BlockTime()) {
		return nil
	}

	return k.Compress(ctx, acc)
}
