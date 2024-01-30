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

// Parameter store keys
var (
	KeyCompanyAccounts = []byte("CompanyAccounts")
	KeyTransitionCost  = []byte("TransitionCost")
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
	return ca.ForSubscription == bech32
}

func (ca CompanyAccounts) GetForSubscription() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ca.ForSubscription)
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
		paramTypes.NewParamSetPair(KeyTransitionCost, &p.TransitionPrice, validateUint64),
	}
}

func (p Params) Validate() error {
	if err := validateCompanyAccounts(p.CompanyAccounts); err != nil {
		return err
	}
	return nil
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return Params{
		TransitionPrice: DefaultTransitionPrice,
	}
}

func validateCompanyAccounts(i interface{}) error {
	ca, ok := i.(CompanyAccounts)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if _, err := sdk.AccAddressFromBech32(ca.ForSubscription); err != nil {
		return errors.Wrap(err, "cannot parse for_subscription account address")
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
