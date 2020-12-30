package keeper

import (
	"encoding/base64"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/arterynetwork/artr/x/storage/types"
)

func (k Keeper) ExportLimits(ctx sdk.Context) []types.Volume {
	var result []types.Volume
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, limitPrefix)
	defer it.Close()
	offset := len(limitPrefix) + len(auth.AddressStoreKeyPrefix)
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key()[offset:])
		volume := binary.BigEndian.Uint64(it.Value())
		result = append(result, types.Volume{
			Account: acc,
			Volume:  volume,
		})
	}
	return result
}

func (k Keeper) ExportCurrent(ctx sdk.Context) []types.Volume {
	var result []types.Volume
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, currentPrefix)
	defer it.Close()
	offset := len(currentPrefix) + len(auth.AddressStoreKeyPrefix)
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key()[offset:])
		volume := binary.BigEndian.Uint64(it.Value())
		result = append(result, types.Volume{
			Account: acc,
			Volume:  volume,
		})
	}
	return result
}

func (k Keeper) ExportData(ctx sdk.Context) []types.Data {
	var result []types.Data
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, dirPrefix)
	defer it.Close()
	offset := len(dirPrefix) + len(auth.AddressStoreKeyPrefix)
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key()[offset:])
		result = append(result, types.Data{
			Account: acc,
			Base64:  base64.StdEncoding.EncodeToString(it.Value()),
		})
	}
	return result
}
