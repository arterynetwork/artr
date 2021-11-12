package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
)

const freeImStorageGb = 5

func (p Profile) IsActive(ctx sdk.Context) bool {
	return p.ActiveUntil != nil && p.ActiveUntil.After(ctx.BlockTime())
}

func (p Profile) IsExtraImStorageActive(ctx sdk.Context) bool {
	return p.ImLimitExtra != 0 && p.ExtraImUntil != nil && p.ExtraImUntil.After(ctx.BlockTime())
}

func (p Profile) ImLimitTotal(ctx sdk.Context) uint64 {
	var extra uint64
	if p.IsExtraImStorageActive(ctx) {
		extra = p.ImLimitExtra
	}
	return (freeImStorageGb + extra) * util.GBSize
}

func NewProfile(activeUntil time.Time, autoPay, noding, storage, validator, vpn bool, nickname string, cardNo uint64) Profile {
	profile := Profile{
		AutoPay:     autoPay,
		ActiveUntil: &time.Time{},
		Noding:      noding,
		Storage:     storage,
		Validator:   validator,
		Vpn:         vpn,
		Nickname:    nickname,
		CardNumber:  cardNo,
	}
	*profile.ActiveUntil = activeUntil
	return profile
}
