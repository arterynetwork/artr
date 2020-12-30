package types

import (
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

	DefaultMinimalPercent      = 21
	DefaultThousandPlusPercent = 24
	DefaultTenKPlusPercent     = 27
	DefaultHundredKPlusPercent = 30
)

// Parameter store keys
var (
	KeyPercentage = []byte("Percentage")
)

// ParamKeyTable for delegating module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

type Percentage struct {
	Minimal      int `json:"minimal" yaml:"minimal"`
	ThousandPlus int `json:"thousand_plus" yaml:"thousand_plus"`
	TenKPlus     int `json:"ten_k_plus" yaml:"ten_k_plus"`
	HundredKPlus int `json:"hundred_k_plus" yaml:"hundred_k_plus"`
}

func NewPercentage(minimal int, oneK int, tenK int, hundredK int) Percentage {
	return Percentage{
		Minimal:      minimal,
		ThousandPlus: oneK,
		TenKPlus:     tenK,
		HundredKPlus: hundredK,
	}
}

func (p Percentage) Validate() error { return validatePercentage(p) }

// Params - used for initializing default parameter for delegating at genesis
type Params struct {
	Percentage Percentage `json:"percentage" yaml:"percentage"`
}

// NewParams creates a new Params object
func NewParams(percentage Percentage) Params {
	return Params{
		Percentage: percentage,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyPercentage, &p.Percentage, validatePercentage),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		NewPercentage(
			DefaultMinimalPercent,
			DefaultThousandPlusPercent,
			DefaultTenKPlusPercent,
			DefaultHundredKPlusPercent,
		),
	)
}

func (p Params) Validate() error {
	if err := validatePercentage(p.Percentage); err != nil {
		return err
	}
	return nil
}

func validatePercentage(i interface{}) error {
	p, ok := i.(Percentage)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if p.Minimal <= 0 {
		return errors.New("minimal percent is non-positive")
	}
	if p.ThousandPlus <= 0 {
		return errors.New("1000+ percent is non-positive")
	}
	if p.TenKPlus <= 0 {
		return errors.New("10k+ percent is non-positive")
	}
	if p.HundredKPlus <= 0 {
		return errors.New("100k+ percent is non-positive")
	}
	if p.Minimal > p.ThousandPlus {
		return errors.New("minimal percent is reater than 1000+ one")
	}
	if p.ThousandPlus > p.TenKPlus {
		return errors.New("1000+ percent is greater than 10k+ one")
	}
	if p.TenKPlus > p.HundredKPlus {
		return errors.New("10k+ percent is greater than 100k+ one")
	}
	return nil
}
