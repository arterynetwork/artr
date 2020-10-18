package types

// referral module event types
const (
	EventTypeStatusUpdated           = "status_updated"
	EventTypeStatusWillBeDowngraded  = "status_will_be_downgraded"
	EventTypeStatusDowngradeCanceled = "status_downgrade_canceled"
	EventTypeCompression             = "compression"
	EventTypeStatusBonus             = "status_bonus"

	AttributeKeyAddress      = "address"
	AttributeKeyBlockHeight  = "block_height"
	AttributeKeyStatusBefore = "status_before"
	AttributeKeyStatusAfter  = "status_after"
	AttributeKeyReferrer     = "referrer"
	AttributeKeyReferrals    = "referrals"
	AttributeKeyAmount       = "amount"

	AttributeValueCategory = ModuleName
)
