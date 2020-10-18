package types

// noding module event types
const (
	EventTypeValidatorBanished = "validator_banished"
	EventTypeValidatorJailed   = "validator_jailed"
	EventTypeValidatorWarning  = "validator_warning"
	EventTypeValidatorBanned   = "validator_banned"

	AttributeKeyAccountAddress = "account_address"
	AttributeKeyReason         = "reason"
	AttributeKeyEvidences      = "evidences"

	AttributeValueNotEnoughStatus     = "not_enough_status"
	AttributeValueNotEnoughDelegation = "not_enough_delegation"
)
