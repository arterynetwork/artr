package types

const (
	// ModuleName is the name of the module
	ModuleName = "referral"

	// StoreKey to be used when creating the KVStore
	StoreKey      = ModuleName
	IndexStoreKey = ModuleName + "-index"

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)
