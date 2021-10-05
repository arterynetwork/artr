package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/schedule/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, params types.Params, tasks []types.Task) {
	k.setParams(ctx, params)
	for _, t := range tasks {
		k.scheduleTask(ctx, t)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) (types.Params, []types.Task) {
	params := k.GetParams(ctx)

	var tasks []types.Task
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var schedule types.Schedule
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &schedule)
		tasks = append(tasks, schedule.Tasks...)
	}
	return params, tasks
}
