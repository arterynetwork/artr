package types

const (
	// ModuleName is the name of the module
	ModuleName = "voting"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName

	HookName = ModuleName + "/complete"
)

var (
	KeyGovernment       = []byte("government")
	KeyAgreedMembers    = []byte("agreed")
	KeyDisagreedMembers = []byte("disagreed")
	KeyCurrentVote      = []byte("current_vote")
	KeyTotalVotes       = []byte("total_votes")
	KeyTotalAgreed      = []byte("total_agreed")
	KeyTotalDisagreed   = []byte("total_disagreed")
	KeyStartBlock       = []byte("start_block")
	KeyHistoryPrefix    = []byte("h")
)
