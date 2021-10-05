package types

const (
	// ModuleName is the name of the module
	ModuleName = "delegating"

	// MainStoreKey to be used when creating the KVStore
	MainStoreKey    = ModuleName
	ClusterStoreKey = MainStoreKey + "-clusters"

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName

	RevokeHookName = "delegating/revoke"
	AccrueHookName = "delegating/accrue"
)
