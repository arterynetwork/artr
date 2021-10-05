package types

const (
	// ModuleName is the name of the module
	ModuleName = "earning"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName

	// VpnCollectorName is the root string for an account address for Artery VPN tariff payment collection
	VpnCollectorName = "vpn"

	// StorageCollectorName is the root string for an account address for Artery Storage tariff payment collection
	StorageCollectorName = "storage"
)
