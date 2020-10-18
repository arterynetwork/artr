package types

// subscription module event types
const (
	EventTypePaySubscription = "pay_subscription"
	EventTypePayVPN          = "pay_vpn"
	EventTypePayStorage      = "pay_storage"
	EventTypeFee             = "subscription_fee"
	EventTypeActivityChange  = "activity_change"
	EventTypeAutoPayFailed   = "autopay_failed"

	AttributeKeyAddress  = "address"
	AttributeKeyExpireAt = "expire_at"
	AttributeKeyLimit    = "limit"
	AttributeKeyAmount   = "amount"
	AttributeKeyNodeFee  = "node_fee"
	AttributeKeyActive   = "active"

	AttributeValueCategory          = ModuleName
	AttributeValueKeyActiveActive   = "active"
	AttributeVAlueKeyActiveInactive = "inactive"
)
