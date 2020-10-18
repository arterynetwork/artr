package types

import "fmt"

// Query endpoints supported by the voting querier
const (
	QueryParams     = "params"
	QueryGovernment = "government"
	QueryCurrent    = "current"
	QueryStatus     = "status"
	QueryHistory    = "history"
)

type QueryHistoryParams struct {
	Limit int32 `json:"limit" yaml:"limit"`
	Page  int32 `json:"page" yaml:"page"`
}

func (q QueryHistoryParams) String() string {
	return fmt.Sprintf("Limit: %d\nPage: %d\n", q.Limit, q.Page)
}

type QueryGovernmentRes Government

var NewQueryGovernmentRes = NewGovernment

type QueryCurrentRes Proposal

func NewQueryCurrentRes(proposal Proposal) QueryCurrentRes {
	return QueryCurrentRes(proposal)
}

type QueryStatusRes struct {
	Proposal   Proposal   `json:"proposal" yaml:"proposal"`
	Government Government `json:"government" yaml:"govenment"`
	Agreed     Government `json:"agreed" yaml:"agreed"`
	Disagreed  Government `json:"disagreed" yaml:"disagreed"`
}

func NewQueryStatusRes(
	proposal Proposal,
	government Government,
	agreed Government,
	disagreed Government,
) QueryStatusRes {
	return QueryStatusRes{
		Proposal:   proposal,
		Government: government,
		Agreed:     agreed,
		Disagreed:  disagreed,
	}
}
