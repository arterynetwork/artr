package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryBalanceParams defines the params for querying an account balance.
type QueryBalanceParams struct {
	Address sdk.AccAddress
}

// NewQueryBalanceParams creates a new instance of QueryBalanceParams.
func NewQueryBalanceParams(addr sdk.AccAddress) QueryBalanceParams {
	return QueryBalanceParams{Address: addr}
}

type QueryResParams struct {
	SendEnabled bool  `json:"send_enabled" yaml:"send_enabled"`
	MinSend     int64 `json:"min_send" yaml:"min_send"`
}

func NewQueryResParams(sendEnabled bool, minSend int64) QueryResParams {
	return QueryResParams{
		SendEnabled: sendEnabled,
		MinSend:     minSend,
	}
}
