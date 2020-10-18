package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the profile querier
const (
	QueryProfile                    = "profile"
	QueryAccountAddressByNickname   = "query_account_address_by_nickname"
	QueryAccountAddressByCardNumber = "query_account_address_by_card_number"
	QueryCreators                   = "query_creators"
)

type QueryResProfile struct {
	Profile Profile `json:"profile" yaml:"profile"`
}

func (q QueryResProfile) String() string {
	return q.Profile.String()
}

// QueryProfileParams defines the params for querying an account balance.
type QueryProfileParams struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

// NewQueryProfileParams creates a new instance of QueryProfileParams.
func NewQueryProfileParams(addr sdk.AccAddress) QueryProfileParams {
	return QueryProfileParams{Address: addr}
}

type QueryResAccountBy struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

func (q QueryResAccountBy) String() string {
	return q.Address.String()
}

// QueryProfileParams defines the params for querying an account number by nickname.
type QueryAccountByNicknameParams struct {
	Nickname string `json:"nickname" yaml:"nickname"`
}

// NewQueryAccountByNicknameParams creates a new instance of QueryBalanceParams.
func NewQueryAccountByNicknameParams(nickname string) QueryAccountByNicknameParams {
	return QueryAccountByNicknameParams{Nickname: nickname}
}

// QueryProfileParams defines the params for querying an account number by nickname.
type QueryAccountByCardNumberParams struct {
	CardNumber uint64 `json:"card_number" yaml:"card_number"`
}

// NewQueryAccountByNicknameParams creates a new instance of QueryBalanceParams.
func NewQueryAccountByCardNumberParams(cardNumber uint64) QueryAccountByCardNumberParams {
	return QueryAccountByCardNumberParams{CardNumber: cardNumber}
}

type QueryCreatorsParams struct{}

func (q QueryCreatorsParams) String() string {
	return ""
}

type QueryCreatorsRes struct {
	Creators []sdk.AccAddress `json:"creators" yaml:"creators"`
}

func (q QueryCreatorsRes) String() string {
	return fmt.Sprintln(q.Creators)
}

func NewQueryCreatorsRes(creators []sdk.AccAddress) QueryCreatorsRes {
	return QueryCreatorsRes{Creators: creators}
}
