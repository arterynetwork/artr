package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Query endpoints supported by the delegating querier
const (
	QueryParams       = "params"
	QueryRevoking     = "revoking"
	QueryAccumulation = "accum"
)

func (req RevokingRequest) GetAccAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.AccAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (req AccumulationRequest) GetAccAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.AccAddress)
	if err != nil {
		panic(err)
	}
	return addr
}
