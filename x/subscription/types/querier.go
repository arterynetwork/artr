package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the subscription querier
const (
	QueryActivityInfo = "info"
	QueryPrices       = "prices"
)

type QueryActivityParams struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

func (params QueryActivityParams) String() string {
	return params.Address.String()
}

func NewQueryActivityInfoParams(addr sdk.AccAddress) QueryActivityParams {
	return QueryActivityParams{
		Address: addr,
	}
}

type QueryActivityRes struct {
	ExpireAt int64 `json:"expire_at" yaml:"expire_at"`
	Active   bool  `json:"active" yaml:"active"`
	Current  int64 `json:"current" yaml:"current"`
}

func (res QueryActivityRes) String() string {
	return fmt.Sprintf(
		"Active: %t\n"+
			"ExpireAt: %d\n"+
			"Current: %d\n",
		res.Active,
		res.ExpireAt,
		res.Current,
	)
}

func NewQueryActivityInfoRes(info ActivityInfo, current int64) QueryActivityRes {
	return QueryActivityRes{
		ExpireAt: info.ExpireAt,
		Active:   info.Active,
		Current:  current,
	}
}

type QueryPricesRes struct {
	Subscription int64 `json:"subscription" yaml:"subscription"`
	VPN          int64 `json:"vpn" yaml:"vpm"`
	Storage      int64 `json:"storage" yaml:"storage"`
	Course       int64 `json:"course" yaml:"course"`
	StorageGb    int32 `json:"storage_base" yaml:"storage_base"`
	VPNGb        int32 `json:"vpn_base" yaml:"vpn_base"`
}

func (res QueryPricesRes) String() string {
	return fmt.Sprintf(
		"Subscription: %d\n"+
			"VPN: %d\n"+
			"Storage: %d\n"+
			"StorageGB: %d\n"+
			"VPNGb: %d\n",
		res.Subscription,
		res.VPN,
		res.Storage,
		res.StorageGb,
		res.VPNGb,
	)
}

/*
Below you will be able how to set your own queries:


// QueryResList Queries Result Payload for a query
type QueryResList []string

// implement fmt.Stringer
func (n QueryResList) String() string {
	return strings.Join(n[:], "\n")
}

*/
