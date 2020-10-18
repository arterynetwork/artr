package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/x/vpn/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the vpn store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	paramspace types.ParamSubspace
}

// NewKeeper creates a vpn keeper
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

// Get returns the pubkey from the adddress-pubkey relation
func (k Keeper) GetInfo(ctx sdk.Context, addr sdk.AccAddress) (types.VpnInfo, error) {
	store := ctx.KVStore(k.storeKey)
	var item types.VpnInfo
	byteKey := auth.AddressStoreKey(addr)
	bz := store.Get(byteKey)

	if bz == nil {
		return types.VpnInfo{}, nil
	}

	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &item)
	if err != nil {
		return types.VpnInfo{}, err
	}
	return item, nil
}

func (k Keeper) SetInfo(ctx sdk.Context, addr sdk.AccAddress, value types.VpnInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(value)
	store.Set(auth.AddressStoreKey(addr), bz)
}

func (k Keeper) GetLimit(ctx sdk.Context, addr sdk.AccAddress) (int64, error) {
	info, err := k.GetInfo(ctx, addr)

	if err != nil {
		return 0, err
	}

	return info.Limit, nil
}

func (k Keeper) GetCurrent(ctx sdk.Context, addr sdk.AccAddress) (int64, error) {
	info, err := k.GetInfo(ctx, addr)

	if err != nil {
		return 0, err
	}

	return info.Current, nil
}

func (k Keeper) SetLimit(ctx sdk.Context, addr sdk.AccAddress, value int64) {
	info, err := k.GetInfo(ctx, addr)
	if err != nil {
		info = types.VpnInfo{Limit: value}
	} else {
		info.Limit = value
	}

	k.SetInfo(ctx, addr, info)
}

func (k Keeper) SetCurrent(ctx sdk.Context, addr sdk.AccAddress, value int64) {
	info, err := k.GetInfo(ctx, addr)
	if err != nil {
		info = types.VpnInfo{Current: value}
	} else {
		info.Current = value
	}

	k.SetInfo(ctx, addr, info)
}

func (k Keeper) AddLimit(ctx sdk.Context, addr sdk.AccAddress, value int64) (int64, error) {
	current, err := k.GetInfo(ctx, addr)

	if err != nil {
		return 0, err
	}

	current.Limit += value

	k.SetInfo(ctx, addr, current)

	return current.Limit, nil
}

func (k Keeper) IterateInfo(ctx sdk.Context, cb func(info types.VpnInfo, addr sdk.AccAddress) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(nil, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		account := k.decodeVpnInfo(iterator.Value())
		addr := sdk.AccAddress(iterator.Key()[1:])

		if cb(account, addr) {
			break
		}
	}
}

func (k Keeper) decodeVpnInfo(bz []byte) (info types.VpnInfo) {
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &info)
	if err != nil {
		panic(err)
	}
	return
}
