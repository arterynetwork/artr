package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/delegating/types"
)

func (k Keeper) InitClusters(ctx sdk.Context, clusters []types.Cluster) {
	mainStore    := ctx.KVStore(k.mainStoreKey)
	clusterStore := ctx.KVStore(k.clusterStoreKey)
	for _, cluster := range clusters {
		key := make([]byte, 2)
		binary.BigEndian.PutUint16(key, cluster.Modulo)
		value :=  k.cdc.MustMarshalBinaryLengthPrefixed(cluster.Accounts)
		clusterStore.Set(key, value)

		value = k.cdc.MustMarshalBinaryLengthPrefixed(types.Record{Cluster: int64(cluster.Modulo)})
		for _, account := range cluster.Accounts {
			mainStore.Set(account, value)
		}
	}
}

func (k Keeper) ExportClusters(ctx sdk.Context) []types.Cluster {
	var result []types.Cluster
	store := ctx.KVStore(k.clusterStoreKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		key := binary.BigEndian.Uint16(it.Key())
		var value []sdk.AccAddress
		k.cdc.MustUnmarshalBinaryLengthPrefixed(it.Value(), &value)
		result = append(result, types.Cluster{
			Modulo:   key,
			Accounts: value,
		})
	}
	return result
}

func (k Keeper) InitRevokeRequests(ctx sdk.Context, revoking []types.Revoke) {
	store := ctx.KVStore(k.mainStoreKey)
	for _, req := range revoking {
		byteKey := []byte(req.Account)

		var item types.Record
		if store.Has(byteKey) {
			byteItem := store.Get(byteKey)
			k.cdc.MustUnmarshalBinaryLengthPrefixed(byteItem, &item)
		} else {
			item = types.NewRecord()
		}

		item.Requests = append(item.Requests, types.RevokeRequest{
			HeightToImplementAt: req.Height,
			MicroCoins:          sdk.NewInt(req.Amount),
		})
		store.Set(byteKey, k.cdc.MustMarshalBinaryLengthPrefixed(item))
	}
}

func (k Keeper) ExportRevokeRequests(ctx sdk.Context) []types.Revoke {
	var result []types.Revoke
	store := ctx.KVStore(k.mainStoreKey)
	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		acc := sdk.AccAddress(it.Key())
		var r types.Record
		k.cdc.MustUnmarshalBinaryLengthPrefixed(it.Value(), &r)
		for _, req := range r.Requests {
			result = append(result, types.Revoke{
				Account: acc,
				Amount:  req.MicroCoins.Int64(),
				Height:  req.HeightToImplementAt,
			})
		}
	}
	return result
}