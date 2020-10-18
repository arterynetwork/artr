package types

import (
	"fmt"
	"strings"
)

// Query endpoints supported by the delegating querier
const (
	QueryRevoking     = "revoking"
	QueryAccumulation = "accum"
)

type QueryResRevoking []RevokeRequest

type QueryResAccumulation struct {
	StartHeight   int64 `json:"start_height"`
	EndHeight     int64 `json:"end_height"`
	Percent       int   `json:"int"`
	TotalUartrs   int64 `json:"total_uartrs"`
	CurrentUartrs int64 `json:"current_uartrs"`
}

func (x QueryResRevoking) String() string {
	if x == nil {
		return "none"
	}
	sb := strings.Builder{}
	for _, q := range x {
		_, err := sb.WriteString(fmt.Sprintf("%d uartr at height %d\n", q.MicroCoins, q.HeightToImplementAt))
		if err != nil {
			panic(err)
		}
	}
	return sb.String()[:sb.Len()-1]
}
