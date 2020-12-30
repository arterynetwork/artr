package types

import (
	"github.com/arterynetwork/artr/util"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Status int

const (
	Lucky            Status = 1
	Leader           Status = 2
	Master           Status = 3
	Champion         Status = 4
	Businessman      Status = 5
	Professional     Status = 6
	TopLeader        Status = 7
	Hero             Status = 8
	AbsoluteChampion Status = 9

	MinimumStatus = Lucky
	MaximumStatus = AbsoluteChampion
)

func (s Status) LinesOpened() int {
	switch s {
	case Lucky:
		return 2
	case Leader:
		return 4
	case Master:
		return 6
	default:
		return 10
	}
}

func (s Status) String() string {
	switch s {
	case Lucky:
		return "1 (Lucky)"
	case Leader:
		return "2 (Leader)"
	case Master:
		return "3 (Master)"
	case Champion:
		return "4 (Champion)"
	case Businessman:
		return "5 (Businessman)"
	case Professional:
		return "6 (Professional)"
	case TopLeader:
		return "7 (Top Leader)"
	case Hero:
		return "8 (Hero)"
	case AbsoluteChampion:
		return "9 (Absolute Champion)"
	default:
		return fmt.Sprintf("%d (???)", int(s))
	}
}

type R struct {
	// Status - account status (1 "Lucky" – 9 "Absolute Champion").
	Status Status `json:"status"`

	// Block height at that the account status downgrade is scheduled. -1 for never.
	StatusDowngradeAt int64 `json:"status_downgrade_at"`

	// Referrer - parent, account just above this one.
	Referrer sdk.AccAddress `json:"referrer"`

	// Referrals - children, accounts just below this one.
	Referrals []sdk.AccAddress `json:"referrals"`

	// Coins - total amount of coins (delegated and not) per level:
	// [0] is its own coins, [1] is its children's coins total and so on
	Coins [11]sdk.Int `json:"coins"`

	// Delegated - total amount of delegated coins per level:
	// [0] - delegated by itself, [1] - delegated by children and so on
	Delegated [11]sdk.Int `json:"delegated"`

	// Active - does the account keeper have a paid subscription.
	Active bool `json:"active"`

	// ActiveReferralsCount - count of referral per level.
	// ActiveReferralsCount[1] is in essence just len(Referrals), but only those referrals who have Active == true are counted.
	// ActiveReferralsCount[2] is a total count of all active referral of all account's referrals (whether active of not).
	// And so on. ActiveReferralsCount[0] represents account itself. It must be equal 1 if account is active, and 0 if not.
	ActiveReferralsCount [11]int `json:"active_referrals_count"`

	// CompressionAt - block height, at that compression is scheduled. -1 for never.
	CompressionAt int64 `json:"compression_at"`

	// Transition - a new referrer, the user wishes to be moved under. It should be nil unless the user requested a
	// transition and that transition's waiting for a current referrer's affirmation.
	Transition sdk.AccAddress `json:"transition,omitempty"`
}

func NewR(referrer sdk.AccAddress, coins sdk.Int, delegated sdk.Int) R {
	zero := sdk.ZeroInt()
	return R{
		Status:               Lucky,
		StatusDowngradeAt:    -1,
		Referrer:             referrer,
		Referrals:            nil,
		Coins:                [11]sdk.Int{coins, zero, zero, zero, zero, zero, zero, zero, zero, zero, zero},
		Delegated:            [11]sdk.Int{delegated, zero, zero, zero, zero, zero, zero, zero, zero, zero, zero},
		Active:               false,
		ActiveReferralsCount: [11]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		CompressionAt:        -1,
	}
}

func (r R) CoinsAtLevelsUpTo(n int) sdk.Int {
	result := sdk.NewInt(0)
	for i := 0; i <= n; i++ {
		result = result.Add(r.Coins[i])
	}
	return result
}

func (r R) DelegatedAtLevelsUpTo(n int) sdk.Int {
	result := sdk.NewInt(0)
	for i := 0; i <= n; i++ {
		result = result.Add(r.Delegated[i])
	}
	return result
}

type ReferralFee struct {
	Beneficiary sdk.AccAddress `json:"beneficiary"`
	Ratio       util.Fraction  `json:"ratio"`
}

type StatusCheckResult struct {
	Overall  bool            `json:"overall"`
	Criteria map[string]bool `json:"criteria"`
}

func NewStatusCheckResult() StatusCheckResult {
	return StatusCheckResult{
		Overall:  true,
		Criteria: make(map[string]bool, 2),
	}
}
