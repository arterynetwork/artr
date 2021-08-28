package keeper

import (
	"encoding/binary"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/noding/types"
)

func (k Keeper) lotteryAddNew(ctx sdk.Context, acc sdk.AccAddress, data *types.Info) error {
	if data.LotteryNo != 0 {
		return errors.New("already in")
	}

	store := ctx.KVStore(k.indexStoreKey)
	var n uint64

	it := sdk.KVStoreReversePrefixIterator(store, IdxPrefixLotteryQueue)
	if it.Valid() {
		n = binary.BigEndian.Uint64(it.Key()[len(IdxPrefixLotteryQueue):]) + 1
	} else {
		n = 1 // let's leave 0 for "no value"
	}
	it.Close()

	data.LotteryNo = n
	store.Set(k.lotteryKey(n), acc)
	return nil
}

func (k Keeper) lotteryExclude(ctx sdk.Context, data *types.Info) error {
	if data.LotteryNo == 0 {
		return errors.New("already out")
	}

	store := ctx.KVStore(k.indexStoreKey)
	key := k.lotteryKey(data.LotteryNo)
	store.Delete(key)

	data.LotteryNo = 0
	return nil
}

func (k Keeper) lotteryLastNo(ctx sdk.Context, count int) uint64 {
	store := ctx.KVStore(k.indexStoreKey)
	key := make([]byte, len(IdxPrefixLotteryQueue)+8)
	it := sdk.KVStorePrefixIterator(store, IdxPrefixLotteryQueue)
	for i := 0; i < count; i++ {
		if !it.Valid() {
			break
		}
		copy(key, it.Key())
		it.Next()
	}
	it.Close()
	return binary.BigEndian.Uint64(key[len(IdxPrefixLotteryQueue):])
}

func (k Keeper) lotteryDownshift(ctx sdk.Context, account sdk.AccAddress, data *types.Info) error {
	if err := k.lotteryExclude(ctx, data); err != nil {
		return errors.Wrap(err, "cannot unassign current number")
	}
	if err := k.lotteryAddNew(ctx, account, data); err != nil {
		return errors.Wrap(err, "cannot assign new number")
	}
	return nil
}

func (k Keeper) lotteryKey(n uint64) []byte {
	key := make([]byte, len(IdxPrefixLotteryQueue)+8)
	copy(key, IdxPrefixLotteryQueue)
	binary.BigEndian.PutUint64(key[len(IdxPrefixLotteryQueue):], n)
	return key
}
