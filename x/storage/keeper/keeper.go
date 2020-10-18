package keeper

import (
	"encoding/binary"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	currentPrefix = []byte{0x00}
	limitPrefix   = []byte{0x01}
	dirPrefix     = []byte{0x02}
)

// Keeper of the storage store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	paramspace types.ParamSubspace
}

// NewKeeper creates a storage keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramspace types.ParamSubspace) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		paramspace: paramspace.WithKeyTable(types.ParamKeyTable()),
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddLimit(ctx sdk.Context, addr sdk.AccAddress, volume int64) (int64, error) {
	limit := k.GetLimit(ctx, addr)
	limit += volume
	k.SetLimit(ctx, addr, limit)

	return limit, nil
}

func (k Keeper) SetLimit(ctx sdk.Context, addr sdk.AccAddress, volume int64) {
	store := ctx.KVStore(k.storeKey)
	byteKey := append(limitPrefix, auth.AddressStoreKey(addr)...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(volume))

	store.Set(byteKey, bz)
}

func (k Keeper) GetLimit(ctx sdk.Context, addr sdk.AccAddress) int64 {
	store := ctx.KVStore(k.storeKey)
	byteKey := append(limitPrefix, auth.AddressStoreKey(addr)...)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0
	}

	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) AddCurrent(ctx sdk.Context, addr sdk.AccAddress, volume int64) (int64, error) {
	limit := k.GetLimit(ctx, addr)
	limit += volume
	k.SetLimit(ctx, addr, limit)

	return limit, nil
}

func (k Keeper) SetCurrent(ctx sdk.Context, addr sdk.AccAddress, volume int64) {
	store := ctx.KVStore(k.storeKey)
	byteKey := append(currentPrefix, auth.AddressStoreKey(addr)...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(volume))

	store.Set(byteKey, bz)
}

func (k Keeper) GetCurrent(ctx sdk.Context, addr sdk.AccAddress) int64 {
	store := ctx.KVStore(k.storeKey)
	byteKey := append(currentPrefix, auth.AddressStoreKey(addr)...)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0
	}

	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) SetData(ctx sdk.Context, addr sdk.AccAddress, bz []byte) {
	store := ctx.KVStore(k.storeKey)
	byteKey := append(dirPrefix, auth.AddressStoreKey(addr)...)
	store.Set(byteKey, bz)
}

func (k Keeper) GetData(ctx sdk.Context, addr sdk.AccAddress) []byte {
	store := ctx.KVStore(k.storeKey)
	byteKey := append(dirPrefix, auth.AddressStoreKey(addr)...)
	return store.Get(byteKey)
}
