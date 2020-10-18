package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type RevokeRequest struct {
	HeightToImplementAt int64   `json:"height"`
	MicroCoins          sdk.Int `json:"ucoins"`
}

type Record struct {
	Cluster  int64           `json:"cluster"`
	Requests []RevokeRequest `json:"requests"`
}

func NewRecord() Record {
	return Record{Cluster: -1}
}

func (x Record) IsEmpty() bool {
	return x.Requests == nil && x.Cluster < 0
}
