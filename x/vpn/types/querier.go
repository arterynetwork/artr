package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the vpn querier
const (
	QueryVpnState   = "query_state"
	QueryVpnLimit   = "query_limit"
	QueryVpnCurrent = "query_current"
)

type QueryResState struct {
	State VpnInfo `json:"vpn_info" yaml:"vpn_info"`
}

func (res QueryResState) String() string {
	return res.State.String()
}

type QueryResLimit struct {
	Limit int64 `json:"limit" yaml:"limit"`
}

func (res QueryResLimit) String() string {
	return fmt.Sprintf("%d", res.Limit)
}

type QueryResCurrent struct {
	Current int64 `json:"current" yaml:"current"`
}

func (res QueryResCurrent) String() string {
	return fmt.Sprintf("%d", res.Current)
}

type QueryVpnParams struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

func (params QueryVpnParams) String() string {
	return params.Address.String()
}

func NewQueryVpnParams(addr sdk.AccAddress) QueryVpnParams {
	return QueryVpnParams{addr}
}
