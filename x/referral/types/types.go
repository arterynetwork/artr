package types

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
)

func ParseStatus(name string) (Status, error) {
	if s, ok := Status_value[name]; !ok {
		return STATUS_UNSPECIFIED, fmt.Errorf("cannot parse status from string: %s", name)
	} else {
		return Status(s), nil
	}
}

func (s Status) Validate() error {
	if s < MinimumStatus || s > MaximumStatus || s == HeroDeprecatedStatus {
		return fmt.Errorf("there is no such status: %d", s)
	}
	return nil
}

func (s Status) LinesOpened() int {
	switch s {
	case STATUS_LUCKY:
		return 4
	case STATUS_LEADER:
		return 6
	case STATUS_MASTER:
		return 8
	default:
		return 10
	}
}

const MinimumStatus = STATUS_LUCKY
const MaximumStatus = STATUS_ABSOLUTE_CHAMPION
const HeroDeprecatedStatus = 8

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
	return r.Referrer != "" && (r.Banished || !r.Active && (r.CompressionAt == nil || ctx.BlockTime().After(r.CompressionAt.Add(-sk.OneMonth(ctx)))))
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

func (ca CompanyAccounts) String() string {
	out, _ := yaml.Marshal(ca)
	return string(out)
}

func (ca CompanyAccounts) Validate() error {
	if _, err := sdk.AccAddressFromBech32(ca.ForSubscription); err != nil {
		return errors.Wrap(err, "cannot parse for_subscription account address")
	}
	return nil
}

func (ca CompanyAccounts) GetForSubscription() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.ForSubscription)
	if err != nil {
		panic(err)
	}
	return addr
}

func (na NetworkAward) String() string {
	out, _ := yaml.Marshal(na)
	return string(out)
}

func (na NetworkAward) Validate() error {
	if na.Company.IsNegative() {
		return fmt.Errorf("company award must be non-negative")
	}
	total := na.Company
	for i := 0; i < 10; i++ {
		if na.Network[i].IsNegative() {
			return fmt.Errorf("level %d award must be non-negative", i+1)
		}
		total = total.Add(na.Network[i])
	}
	if total.GTE(util.Percent(100)) {
		return fmt.Errorf("total network award must be less than 100%%")
	}
	return nil
}

type ReferralValidatorFee struct {
	Beneficiary string        `json:"beneficiary" yaml:"beneficiary"`
	Ratio       util.Fraction `json:"ratio" yaml:"ratio"`
}

func (fee ReferralValidatorFee) GetBeneficiary() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(fee.Beneficiary)
	if err != nil {
		panic(err)
	}
	return addr
}
