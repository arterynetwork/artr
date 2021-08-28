package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (p Profile) IsActive(ctx sdk.Context) bool {
	return p.ActiveUntil != nil && p.ActiveUntil.After(ctx.BlockTime())
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
