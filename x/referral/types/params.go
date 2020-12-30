package types

import (
	"github.com/arterynetwork/artr/util"
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace     = ModuleName
	DefaultTransitionCost = 1_000000
)

var (
	DefaultDelegatingAward = NetworkAward{
		Network: [10]util.Fraction{
			util.Percent(5),
			util.Percent(1),
			util.Percent(1),
			util.Percent(2),
			util.Percent(1),
			util.Percent(1),
			util.Percent(1),
			util.Percent(1),
			util.Percent(1),
			util.Permille(5),
		},
		Company: util.Permille(5),
	}

	DefaultSubscriptionAward = NetworkAward{
		Network: [10]util.Fraction{
			util.Percent(15),
			util.Percent(10),
			util.Percent(7),
			util.Percent(7),
			util.Percent(7),
			util.Percent(7),
			util.Percent(7),
			util.Percent(5),
			util.Percent(2),
			util.Percent(2),
		},
		Company: util.Percent(10),
	}
)

// Parameter store keys
var (
	KeyCompanyAccounts   = []byte("CompanyAccounts")
	KeyDelegatingAward   = []byte("DelegatingAward")
	KeySubscriptionAward = []byte("SubscriptionAward")
	KeyTransitionCost    = []byte("TransitionCost")
)

// ParamKeyTable for referral module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

type CompanyAccounts struct {
	TopReferrer     sdk.AccAddress `json:"top_referrer" yaml:"top_referrer"`
	ForSubscription sdk.AccAddress `json:"for_subscription" yaml:"for_subscription"`
	PromoBonuses    sdk.AccAddress `json:"promo_bonuses" yaml:"promo_bonuses"`
	StatusBonuses   sdk.AccAddress `json:"status_bonuses" yaml:"status_bonuses"`
	LeaderBonuses   sdk.AccAddress `json:"leader_bonuses" yaml:"leader_bonuses"`
	ForDelegating   sdk.AccAddress `json:"for_delegating" yaml:"for_delegating"`
}

type NetworkAward struct {
	Network [10]util.Fraction `json:"network" yaml:"network,flow"`
	Company util.Fraction     `json:"company" yaml:"company"`
}

func (na NetworkAward) Validate() error { return validateNetworkAward(na) }

func (ca CompanyAccounts) Contains(acc sdk.AccAddress) bool {
	return !acc.Empty() && (ca.TopReferrer.Equals(acc) ||
		ca.ForSubscription.Equals(acc) ||
		ca.PromoBonuses.Equals(acc) ||
		ca.StatusBonuses.Equals(acc) ||
		ca.LeaderBonuses.Equals(acc) ||
		ca.ForDelegating.Equals(acc))
}

// Params - used for initializing default parameter for referral at genesis
type Params struct {
	CompanyAccounts   CompanyAccounts `json:"company_accounts" yaml:"company_accounts"`
	DelegatingAward   NetworkAward    `json:"delegating_award" yaml:"delegating_award"`
	SubscriptionAward NetworkAward    `json:"subscription_award" yaml:"subscription_award"`
	TransitionCost    uint64          `json:"transition_cost" yaml:"transition_cost"`
}

// NewParams creates a new Params object
func NewParams(ca CompanyAccounts) Params {
	return Params{
		CompanyAccounts: ca,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyCompanyAccounts, &p.CompanyAccounts, validateCompanyAccounts),
		params.NewParamSetPair(KeyDelegatingAward, &p.DelegatingAward, validateNetworkAward),
		params.NewParamSetPair(KeySubscriptionAward, &p.SubscriptionAward, validateNetworkAward),
		params.NewParamSetPair(KeyTransitionCost, &p.TransitionCost, validateUint64),
	}
}

func (p Params) Validate() error {
	if err := validateCompanyAccounts(p.CompanyAccounts); err != nil {
		return err
	}
	if err := validateNetworkAward(p.DelegatingAward); err != nil {
		return nil
	}
	if err := validateNetworkAward(p.SubscriptionAward); err != nil {
		return nil
	}
	return nil
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return Params{
		DelegatingAward:   DefaultDelegatingAward,
		SubscriptionAward: DefaultSubscriptionAward,
		TransitionCost:    DefaultTransitionCost,
	}
}

func validateCompanyAccounts(i interface{}) error {
	ca, ok := i.(CompanyAccounts)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if ca.TopReferrer.Empty() {
		return errors.New("empty company account for referral bonuses excess")
	}
	if ca.ForSubscription.Empty() {
		return errors.New("empty company account for subscription")
	}
	if ca.PromoBonuses.Empty() {
		return errors.New("empty company account for promo bonuses")
	}
	if ca.StatusBonuses.Empty() {
		return errors.New("empty company account for status bonuses")
	}
	if ca.LeaderBonuses.Empty() {
		return errors.New("empty company account for leader bonuses")
	}
	if ca.ForDelegating.Empty() {
		return errors.New("empty company account for delegating")
	}

	return nil
}

func validateNetworkAward(i interface{}) error {
	na, ok := i.(NetworkAward)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
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

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type (uint64 expected): %T", i)
	}
	return nil
}
