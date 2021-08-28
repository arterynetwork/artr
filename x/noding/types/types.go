package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewInfo(power int64, pubKey string) *Info {
	return &Info{
		Power:             power,
		Status:            true,
		LastPower:         0,
		PubKey:            pubKey,
		Strokes:           0,
		OkBlocksInRow:     0,
		MissedBlocksInRow: 0,
		Jailed:            false,
		UnjailAt:          0,
		Infractions:       nil,
		BannedForLife:     false,
		Staff:             false,
		ProposedCount:     0,
		JailCount:         0,
	}
}

func (x Info) IsActive() bool {
	return x.Status && !x.Jailed && !x.BannedForLife && x.Power != 0
}

type InfoWithAccount struct {
	Info
	Account sdk.AccAddress
}

func NewInfoWithAccount(acc sdk.AccAddress, info Info) InfoWithAccount {
	return InfoWithAccount{
		Info:    info,
		Account: acc,
	}
}
