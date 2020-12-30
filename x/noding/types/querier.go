package types

// Query endpoints supported by the noding querier
const (
	QueryStatus   = "status"
	QueryInfo     = "info"
	QueryProposer = "proposer"
	QueryAllowed  = "allowed"
	QueryOperator = "operator"
	QueryParams   = "params"

	QueryOperatorFormatHex    = "hex"
	QueryOperatorFormatBech32 = "bech32"
)

type AllowedQueryRes struct {
	Verdict bool   `json:"verdict" yaml:"verdict"`
	Reason  string `json:"reason" yaml:"reason"`
}

func NewAllowedQueryRes(verdict bool, reason string) AllowedQueryRes {
	return AllowedQueryRes{
		Verdict: verdict,
		Reason:  reason,
	}
}
