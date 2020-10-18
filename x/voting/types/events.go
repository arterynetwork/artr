package types

// voting module event types
const (
	EventTypeCreateProposal = "create_proposal"
	EventTypeProposalVote   = "proposal_vote"
	EventTypeProposalEnd    = "proposal_end"

	AttributeKeyAuthor     = "author"
	AttributeKeyTypeCode   = "type_code"
	AttributeKeyAgree      = "agree"
	AttributeKeyAgreed     = "agreed"
	AttributeKeyDisagreed  = "disagreed"
	AttributeKeyGovernment = "government"

	AttributeValueCategory = ModuleName
)
