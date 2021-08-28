package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Query endpoints supported by the profile querier
const (
	QueryProfileByAddress    = "get_by_addr"
	QueryProfileByNickname   = "get_by_nick"
	QueryProfileByCardNumber = "get_by_card"
	QueryParams              = "params"
)

func (req GetByAddressRequest) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		panic(err)
	}
	return addr
}
