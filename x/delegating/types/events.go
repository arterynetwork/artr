package types

// delegating module event types
const (
	EventTypeDelegate      = "delegate"
	EventTypeUndelegate    = "undelegate"
	EventTypeAccrue        = "accrue"
	EventTypeMassiveRevoke = "massive_revoke"

	AttributeKeyAccount          = "account"
	AttributeKeyUcoins           = "ucoins"
	AttributeKeyCommissionTo     = "commission_to"
	AttributeKeyCommissionAmount = "commission_amount"

	AttributeValueCategory = ModuleName
)
