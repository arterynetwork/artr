package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewTimestamps(vpn *time.Time, storage *time.Time) Timestamps {
	return Timestamps{
		Vpn:     vpn,
		Storage: storage,
	}
}

func NewEarner(acc sdk.AccAddress, vpn *time.Time, storage *time.Time) Earner {
	return Earner{
		Account: acc.String(),
		Vpn:     vpn,
		Storage: storage,
	}
}

func (e Earner) GetAccount() sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(e.Account)
	if err != nil {
		panic(err)
	}
	return acc
}

func (e Earner) GetTimestamps() Timestamps {
	return Timestamps{
		Vpn:     e.Vpn,
		Storage: e.Storage,
	}
}
