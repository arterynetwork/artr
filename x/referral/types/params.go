package types

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

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

// NewParams creates a new Params object
func NewParams(ca CompanyAccounts, tp uint64) Params {
	return Params{
		CompanyAccounts: ca,
		TransitionPrice: tp,
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return Params{
		TransitionPrice: DefaultTransitionPrice,
	}
}

func (p Params) String() string {
	out, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
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
	if err := validateUint64(p.TransitionPrice); err != nil {
		return err
	}
	return nil
}

func validateCompanyAccounts(i interface{}) error {
	ca, ok := i.(CompanyAccounts)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if err := ca.Validate(); err != nil {
		return errors.Wrap(err, "invalid parameter:")
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
