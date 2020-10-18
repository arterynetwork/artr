package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Points struct {
	Vpn     int64 `json:"vpn"`
	Storage int64 `json:"storage"`
}

func NewPoints(vpn int64, storage int64) Points {
	return Points{
		Vpn:     vpn,
		Storage: storage,
	}
}

type Earner struct {
	Points
	Account sdk.AccAddress `json:"account"`
}

func NewEarner(acc sdk.AccAddress, vpn int64, storage int64) Earner {
	return Earner{
		Points:  NewPoints(vpn, storage),
		Account: acc,
	}
}
