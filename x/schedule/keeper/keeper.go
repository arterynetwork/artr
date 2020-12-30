package keeper

import (
	"encoding/binary"
	"fmt"
	//authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"

	//"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/x/schedule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the schedule store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	paramspace types.ParamSubspace
	eventHooks map[string]func(ctx sdk.Context, data []byte)
}

// NewKeeper creates a schedule keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramspace types.ParamSubspace) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		paramspace: paramspace.WithKeyTable(types.ParamKeyTable()),
		eventHooks: make(map[string]func(ctx sdk.Context, data []byte)),
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Add event hook
func (k Keeper) AddHook(event string, hook func(ctx sdk.Context, data []byte)) {
	k.eventHooks[event] = hook
}

func (k Keeper) GetTasks(ctx sdk.Context, block uint64) types.Schedule {
	store := ctx.KVStore(k.storeKey)

	blockBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBuf, block)

	var items types.Schedule

	bz := store.Get(blockBuf)

	if bz == nil {
		items = make(types.Schedule, 0)
	} else {
		err := k.cdc.UnmarshalBinaryBare(bz, &items)

		if err != nil {
			panic(err)
		}
	}

	return items
}

// Schedule an event on block block height
func (k Keeper) ScheduleTask(ctx sdk.Context, block uint64, event string, data *[]byte) error {
	store := ctx.KVStore(k.storeKey)

	blockBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBuf, block)

	var items types.Schedule

	bz := store.Get(blockBuf)

	if bz == nil {
		items = make(types.Schedule, 0)
	} else {
		err := k.cdc.UnmarshalBinaryBare(bz, &items)

		if err != nil {
			return err
		}
	}

	items = append(items, types.Task{
		HandlerName: event,
		Data:        *data,
	})

	bz = k.cdc.MustMarshalBinaryBare(items)
	store.Set(blockBuf, bz)

	return nil
}

func (k Keeper) filterTasks(vs types.Schedule, excludeEvent string) types.Schedule {
	vsf := make(types.Schedule, 0)
	for _, v := range vs {
		if v.HandlerName != excludeEvent {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func (k Keeper) DeleteAllTasksOnBlock(ctx sdk.Context, block uint64, event string) {
	store := ctx.KVStore(k.storeKey)

	blockBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBuf, block)

	bz := store.Get(blockBuf)

	if bz == nil {
		return
	}

	var items types.Schedule

	err := k.cdc.UnmarshalBinaryBare(bz, &items)

	if err != nil {
		return
	}

	items = k.filterTasks(items, event)

	bz = k.cdc.MustMarshalBinaryBare(items)

	if len(bz) == 0 {
		store.Delete(blockBuf)
	} else {
		store.Set(blockBuf, bz)
	}
}

// Perfoms a sheduled tasks for block height. Tasks removed from store after completion
func (k Keeper) PerfomSchedule(ctx sdk.Context, block uint64) {
	// We can ignore InitialHeight here, because all performed tasks are removed from KVStore

	store := ctx.KVStore(k.storeKey)

	blockBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBuf, block)

	bz := store.Get(blockBuf)

	if bz == nil {
		return
	}

	var items types.Schedule

	err := k.cdc.UnmarshalBinaryBare(bz, &items)

	if err != nil {
		return
	}

	for _, task := range items {
		hook := k.eventHooks[task.HandlerName]
		if hook != nil {
			performSchedule(ctx, task, hook, k.Logger(ctx))
		}
	}

	store.Delete(blockBuf)
}

func performSchedule(ctx sdk.Context, task types.Task, hook func(ctx sdk.Context, data []byte), logger log.Logger) {
	logger.Debug("perform schedule", "task", task.HandlerName)
	defer func(task string) {
		if err := recover(); err != nil {
			logger.Error("recovered from panic",
				"task", task,
				"error", err,
			)
		}
	}(task.HandlerName)
	hook(ctx, task.Data)
}
