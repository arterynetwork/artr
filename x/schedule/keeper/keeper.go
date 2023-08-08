package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/x/schedule/types"
)

// Keeper of the schedule store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryMarshaler
	eventHooks map[string]func(ctx sdk.Context, data []byte, time time.Time)
	paramspace paramtypes.Subspace
}

// NewKeeper creates a schedule keeper
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, paramspace paramtypes.Subspace) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		eventHooks: make(map[string]func(ctx sdk.Context, data []byte, time time.Time)),
		paramspace: paramspace.WithKeyTable(types.ParamKeyTable()),
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Add event hook
func (k Keeper) AddHook(event string, hook func(ctx sdk.Context, data []byte, time time.Time)) {
	k.eventHooks[event] = hook
}

func (k Keeper) GetTasks(ctx sdk.Context, since, to time.Time) []types.Task {
	var (
		store   = ctx.KVStore(k.storeKey)
		items   []types.Task
		sinceBz = Key(since)
		toBz    = Key(to)
		it      = store.Iterator(sinceBz, toBz)
	)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var sch types.Schedule
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &sch)
		items = append(items, sch.Tasks...)
	}
	return items
}

// Schedule an event on block block height
func (k Keeper) ScheduleTask(ctx sdk.Context, time time.Time, event string, data []byte) {
	k.scheduleTask(ctx, types.Task{Time: time, HandlerName: event, Data: data})
}
func (k Keeper) scheduleTask(ctx sdk.Context, task types.Task) {
	var (
		store = ctx.KVStore(k.storeKey)
		key   = Key(task.Time)
		sch   types.Schedule
	)
	if bz := store.Get(key); bz != nil {
		k.cdc.MustUnmarshalBinaryBare(bz, &sch)
	}

	sch.Tasks = append(sch.Tasks, task)

	store.Set(key, k.cdc.MustMarshalBinaryBare(&sch))
}

func (k Keeper) DeleteAll(ctx sdk.Context, time time.Time, event string) {
	k.delete(ctx, time, func(task types.Task) bool {
		return task.HandlerName == event
	})
}
func (k Keeper) Delete(ctx sdk.Context, time time.Time, event string, payload []byte) {
	k.delete(ctx, time, func(task types.Task) bool {
		return task.HandlerName == event && bytes.Equal(task.Data, payload)
	})
}
func (k Keeper) delete(ctx sdk.Context, time time.Time, predicate func(types.Task) bool) {
	store := ctx.KVStore(k.storeKey)
	key := Key(time)
	bz := store.Get(key)

	if bz == nil {
		return
	}

	var items types.Schedule
	k.cdc.MustUnmarshalBinaryBare(bz, &items)

	filtered := make([]types.Task, 0, len(items.Tasks))
	for _, item := range items.Tasks {
		if !predicate(item) {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == 0 {
		store.Delete(key)
	} else {
		store.Set(key, k.cdc.MustMarshalBinaryBare(&types.Schedule{Tasks: filtered}))
	}
}

// PerformSchedule performs scheduled tasks for the block height. Tasks will be removed from the store when they are
// complete.
func (k Keeper) PerformSchedule(ctx sdk.Context) {
	store := cachekv.NewStore(ctx.KVStore(k.storeKey))
	terminator := Key(ctx.BlockTime().Add(time.Nanosecond))
	it := store.Iterator(nil, terminator)

	for ; it.Valid(); it.Next() {
		var sch types.Schedule
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &sch)
		for _, task := range sch.Tasks {
			hook := k.eventHooks[task.HandlerName]
			if hook != nil {
				performSchedule(ctx, task, hook, k.Logger(ctx))
			} else {
				k.Logger(ctx).Error("callback is not registered", "hook", task.HandlerName)
			}
		}
		store.Delete(it.Key())
	}

	it.Close()
	store.Write()
}

func performSchedule(ctx sdk.Context, task types.Task, hook func(ctx sdk.Context, data []byte, time time.Time), logger log.Logger) {
	logger.Debug("perform schedule", "task", task.HandlerName, "time", task.Time)
	defer func(task string) {
		if err := recover(); err != nil {
			logger.Error("recovered from panic",
				"task", task,
				"error", err,
			)
		}
	}(task.HandlerName)
	hook(ctx, task.Data, task.Time)
}

func Key(t time.Time) []byte {
	if t.Year() < 1970 || t.Year() > 2262 {
		panic(errors.Errorf("time is out of range: %s", t.String()))
	}
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(t.UnixNano()))
	return bz
}
