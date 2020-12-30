package types

// earning module event types
const (
	EventTypeStart  = "start-paying-earnings"
	EventTypeFinish = "finish-paying-earnings"
	EventTypeEarn   = "earn"

	AttributeKeyAddress = "address"
	AttributeKeyVpn     = "vpn"
	AttributeKeyStorage = "storage"

	AttributeValueCategory = ModuleName
)
