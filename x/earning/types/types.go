package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewPoints(vpn int64, storage int64) Points {
	return Points{
		Vpn:     vpn,
		Storage: storage,
	}
}

func NewEarner(acc sdk.AccAddress, vpn int64, storage int64) Earner {
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

func (e Earner) GetPoints() Points {
	return Points{
		Vpn:     e.Vpn,
		Storage: e.Storage,
	}
}
