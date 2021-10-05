package types

const (
	// ModuleName is the name of the module
	ModuleName = "noding"

	// StoreKey is to be used when creating the KVStore for module data
	StoreKey    = ModuleName
	IdxStoreKey = StoreKey + "-index"

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)
