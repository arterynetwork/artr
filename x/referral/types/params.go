package types

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
)

// Default parameter namespace
const (
	DefaultParamspace      = ModuleName
	DefaultTransitionPrice = 1_000000
)

var (
	DefaultDelegatingAward = NetworkAward{
		Network: []util.Fraction{
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
		Network: []util.Fraction{
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
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (na NetworkAward) Validate() error { return validateNetworkAward(na) }

func (ca CompanyAccounts) Contains(acc sdk.AccAddress) bool {
	if acc.Empty() {
		return false
	}
	bech32 := acc.String()
	return ca.TopReferrer == bech32 ||
		ca.ForSubscription == bech32 ||
		ca.PromoBonuses == bech32 ||
		ca.StatusBonuses == bech32 ||
		ca.LeaderBonuses == bech32 ||
		ca.ForDelegating == bech32
}

func (ca CompanyAccounts) GetTopReferrer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.TopReferrer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (ca CompanyAccounts) GetForSubscription() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.ForSubscription)
	if err != nil {
		panic(err)
	}
	return addr
}

func (ca CompanyAccounts) GetPromoBonuses() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.PromoBonuses)
	if err != nil {
		panic(err)
	}
	return addr
}

func (ca CompanyAccounts) GetStatusBonuses() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.StatusBonuses)
	if err != nil {
		panic(err)
	}
	return addr
}

func (ca CompanyAccounts) GetLeaderBonuses() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.LeaderBonuses)
	if err != nil {
		panic(err)
	}
	return addr
}

func (ca CompanyAccounts) GetForDelegating() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.ForDelegating)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewParams creates a new Params object
func NewParams(ca CompanyAccounts) Params {
	return Params{
		CompanyAccounts: ca,
	}
}

func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(KeyCompanyAccounts, &p.CompanyAccounts, validateCompanyAccounts),
		paramTypes.NewParamSetPair(KeyDelegatingAward, &p.DelegatingAward, validateNetworkAward),
		paramTypes.NewParamSetPair(KeySubscriptionAward, &p.SubscriptionAward, validateNetworkAward),
		paramTypes.NewParamSetPair(KeyTransitionCost, &p.TransitionPrice, validateUint64),
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
		TransitionPrice:   DefaultTransitionPrice,
	}
}

func validateCompanyAccounts(i interface{}) error {
	ca, ok := i.(CompanyAccounts)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if _, err := sdk.AccAddressFromBech32(ca.TopReferrer); err != nil {
		return errors.Wrap(err, "cannot parse top_referrer account address")
	}
	if _, err := sdk.AccAddressFromBech32(ca.ForSubscription); err != nil {
		return errors.Wrap(err, "cannot parse for_subscription account address")
	}
	if _, err := sdk.AccAddressFromBech32(ca.PromoBonuses); err != nil {
		return errors.Wrap(err, "cannot parse promo_bonuses account address")
	}
	if _, err := sdk.AccAddressFromBech32(ca.StatusBonuses); err != nil {
		return errors.Wrap(err, "cannot parse status_bonuses account address")
	}
	if _, err := sdk.AccAddressFromBech32(ca.LeaderBonuses); err != nil {
		return errors.Wrap(err, "cannot parse leader_bonuses account address")
	}
	if _, err := sdk.AccAddressFromBech32(ca.ForDelegating); err != nil {
		return errors.Wrap(err, "cannot parse for_delegating account address")
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
