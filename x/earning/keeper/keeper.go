package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/earning/types"
)

// Keeper of the earning store
type Keeper struct {
	cdc            codec.BinaryMarshaler
	storeKey       sdk.StoreKey
	paramspace     types.ParamSubspace
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	scheduleKeeper types.ScheduleKeeper
}

// NewKeeper creates a earning keeper
func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	paramspace types.ParamSubspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	scheduleKeeper types.ScheduleKeeper,
) Keeper {
	keeper := Keeper{
		cdc:            cdc,
		storeKey:       key,
		paramspace:     paramspace.WithKeyTable(types.ParamKeyTable()),
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		scheduleKeeper: scheduleKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) IsActiveEarner(ctx sdk.Context, accAddr sdk.AccAddress) (vpn bool, storage bool, err error) {
	vpn = false
	storage = false
	updateTimestamps := false
	if !k.has(ctx, accAddr) {
		return vpn, storage, nil
	}
	timestamps, err := k.get(ctx, accAddr)
	if err != nil {
		return vpn, storage, err
	}
	if timestamps.Vpn != nil {
		if timestamps.Vpn.Add(2 * k.scheduleKeeper.OneDay(ctx)).After(ctx.BlockTime()) {
			vpn = true
		} else {
			timestamps.Vpn = nil
			updateTimestamps = true
		}
	}
	if timestamps.Storage != nil {
		if timestamps.Storage.Add(2 * k.scheduleKeeper.OneDay(ctx)).After(ctx.BlockTime()) {
			storage = true
		} else {
			timestamps.Storage = nil
			updateTimestamps = true
		}
	}
	if timestamps.Vpn == nil && timestamps.Storage == nil {
		k.delete(ctx, accAddr)
	} else if updateTimestamps {
		err := k.set(ctx, accAddr, *timestamps)
		if err != nil {
			return vpn, storage, err
		}
	}
	return vpn, storage, nil
}

//-----------------------------------------------------------------------------------------------------------

func (k Keeper) has(ctx sdk.Context, key sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has([]byte(key))
}

// Get returns the pubkey from the adddress-pubkey relation
func (k Keeper) get(ctx sdk.Context, key sdk.AccAddress) (*types.Timestamps, error) {
	store := ctx.KVStore(k.storeKey)
	var item types.Timestamps
	err := k.cdc.UnmarshalBinaryBare(store.Get([]byte(key)), &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (k Keeper) set(ctx sdk.Context, key sdk.AccAddress, value types.Timestamps) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinaryBare(&value)
	if err != nil {
		return err
	}
	store.Set([]byte(key), bz)
	return nil
}

func (k Keeper) delete(ctx sdk.Context, key sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete([]byte(key))
}

func (k Keeper) clear(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	var keys [][]byte
	it := store.Iterator(nil, nil)
	for ; it.Valid(); it.Next() {
		keys = append(keys, it.Key())
	}
	it.Close()
	for _, key := range keys {
		store.Delete(key)
	}
}
