package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// Query endpoints supported by the noding querier
const (
	QueryStatus     = "status"
	QueryInfo       = "info"
	QueryProposer   = "proposer"
	QueryAllowed    = "allowed"
	QueryOperator   = "operator"
	QueryParams     = "params"
	QuerySwitchedOn = "switched-on"
	QueryState      = "state"

	QueryOperatorFormatHex    = "hex"
	QueryOperatorFormatBech32 = "bech32"
)

func (r GetRequest) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func (r ProposerResponse) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func (r IsAllowedRequest) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func (OperatorRequest) DefaultFormat() OperatorRequest_Format { return OperatorRequest_FORMAT_BECH32 }

func (r OperatorRequest) GetFormat() OperatorRequest_Format {
	if r.Format == OperatorRequest_FORMAT_UNSPECIFIED {
		return r.DefaultFormat()
	} else {
		return r.Format
	}
}

func (r OperatorRequest) GetConsAddress() sdk.ConsAddress {
	var (
		consAddress sdk.ConsAddress
		err         error
	)

	switch format := r.GetFormat(); format {
	case OperatorRequest_FORMAT_BECH32:
		consAddress, err = sdk.ConsAddressFromBech32(r.ConsAddress)
	case OperatorRequest_FORMAT_HEX:
		consAddress, err = sdk.ConsAddressFromHex(r.ConsAddress)
	default:
		panic(errors.Errorf("invalid format %s", format))
	}
	if err != nil {
		panic(err)
	}
	return consAddress
}

func (r OperatorResponse) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func (r SwitchedOnResponse) GetAccounts() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(r.Accounts))
	for i, bech32 := range r.Accounts {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func NewSwitchedOnResponse(list []sdk.AccAddress) *SwitchedOnResponse {
	res := &SwitchedOnResponse{
		Accounts: make([]string, len(list)),
	}
	for i, addr := range list {
		res.Accounts[i] = addr.String()
	}
	return res
}
