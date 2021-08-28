package types

const (
	// ModuleName is the name of the module
	ModuleName = "profile"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// Used when creating account alias KVStore
	AliasStoreKey = ModuleName + "Aliases"

	// Used when creating account card numbers KVStore
	CardStoreKey = ModuleName + "Cards"

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName

	RefreshHookName = ModuleName + "/refresh"
)
