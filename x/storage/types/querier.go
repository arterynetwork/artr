package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the storage querier
const (
	QueryStorageData = "storage_data"
	QueryStorageInfo = "storage_info"
)

type QueryStorageParams struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

func (params QueryStorageParams) String() string {
	return params.Address.String()
}

func NewQueryStorageParams(addr sdk.AccAddress) QueryStorageParams {
	return QueryStorageParams{
		Address: addr,
	}
}

type QueryStorageInfoRes struct {
	Limit   int64 `json:"limit" yaml:"limit"`
	Current int64 `json:"current" yaml:"current"`
}

func (res QueryStorageInfoRes) String() string {
	return fmt.Sprintf(
		"Limit: %d\n"+
			"Current: %d\n",
		res.Limit,
		res.Current,
	)
}

func NewQueryStorageInfoRes(limit int64, current int64) QueryStorageInfoRes {
	return QueryStorageInfoRes{
		Limit:   limit,
		Current: current,
	}
}

type QueryStorageDataRes struct {
	Data string `json:"data" yaml:"data"`
}

func (res QueryStorageDataRes) String() string {
	return fmt.Sprintf(
		"Data: %s\n",
		res.Data,
	)
}

func NewQueryStorageDataRes(data string) QueryStorageDataRes {
	return QueryStorageDataRes{
		Data: data,
	}
}
