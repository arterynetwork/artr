package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
)

func (s Status) LinesOpened() int {
	switch s {
	case STATUS_LUCKY:
		return 2
	case STATUS_LEADER:
		return 4
	case STATUS_MASTER:
		return 6
	default:
		return 10
	}
}

const MinimumStatus = STATUS_LUCKY
const MaximumStatus = STATUS_ABSOLUTE_CHAMPION

func NewInfo(referrer string, coins sdk.Int, delegated sdk.Int) Info {
	zero := sdk.ZeroInt()
	return Info{
		Status:          STATUS_LUCKY,
		Referrer:        referrer,
		Coins:           []sdk.Int{coins, zero, zero, zero, zero, zero, zero, zero, zero, zero, zero},
		Delegated:       []sdk.Int{delegated, zero, zero, zero, zero, zero, zero, zero, zero, zero, zero},
		Active:          false,
		ActiveRefCounts: make([]uint64, 11),
	}
}

func (r Info) CoinsAtLevelsUpTo(n int) sdk.Int {
	result := sdk.NewInt(0)
	for i := 0; i <= n; i++ {
		result = result.Add(r.Coins[i])
	}
	return result
}

func (r Info) DelegatedAtLevelsUpTo(n int) sdk.Int {
	result := sdk.NewInt(0)
	for i := 0; i <= n; i++ {
		result = result.Add(r.Delegated[i])
	}
	return result
}

func (r Info) RegistrationClosed(ctx sdk.Context, sk ScheduleKeeper) bool {
	return r.Banished || !r.Active && (r.CompressionAt == nil || ctx.BlockTime().After(r.CompressionAt.Add(-sk.OneMonth(ctx))))
}

func (r Info) GetReferrer() sdk.AccAddress {
	if r.Referrer == "" {
		return nil
	}
	addr, err := sdk.AccAddressFromBech32(r.Referrer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (r Info) GetTransition() sdk.AccAddress {
	if r.Transition == "" {
		return nil
	}
	addr, err := sdk.AccAddressFromBech32(r.Transition)
	if err != nil {
		panic(err)
	}
	return addr
}

func (r *Info) Normalize() {
	for len(r.Coins) < 11 {
		r.Coins = append(r.Coins, sdk.ZeroInt())
	}
	for len(r.Delegated) < 11 {
		r.Delegated = append(r.Delegated, sdk.ZeroInt())
	}
	for len(r.ActiveRefCounts) < 11 {
		r.ActiveRefCounts = append(r.ActiveRefCounts, uint64(0))
	}
}

func (r Info) IsEmpty() bool {
	return r.Status == STATUS_UNSPECIFIED && !r.Banished
}

type ReferralFee struct {
	Beneficiary string        `json:"beneficiary" yaml:"beneficiary"`
	Ratio       util.Fraction `json:"ratio" yaml:"ratio"`
}

func (fee ReferralFee) GetBeneficiary() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(fee.Beneficiary)
	if err != nil {
		panic(err)
	}
	return addr
}
