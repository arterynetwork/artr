package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/schedule/types"
)

func (k Keeper) InitSchedule(ctx sdk.Context, schedule []types.GenesisSchedule) {
	store := ctx.KVStore(k.storeKey)
	key := make([]byte, 8)
	for _, sch := range schedule {
		binary.BigEndian.PutUint64(key, sch.Height)
		value := k.cdc.MustMarshalBinaryBare(sch.Schedule)
		store.Set(key, value)
	}
}

func (k Keeper) ExportSchedule(ctx sdk.Context) []types.GenesisSchedule {
	var result []types.GenesisSchedule
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		block := binary.BigEndian.Uint64(it.Key())
		var schedule types.Schedule
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &schedule)
		result = append(result, types.GenesisSchedule{
			Schedule: schedule,
			Height:   block,
		})
	}
	return result
}
