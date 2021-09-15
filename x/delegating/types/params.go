package types

import (
	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/util"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

	DefaultMinimalPercent      = 21
	DefaultThousandPlusPercent = 24
	DefaultTenKPlusPercent     = 27
	DefaultHundredKPlusPercent = 30

	DefaultMinDelegate = 1000

	DefaultRevokePeriod = 14 * util.BlocksOneDay
)

// Parameter store keys
var (
	KeyPercentage   = []byte("Percentage")
	KeyMinDelegate  = []byte("MinDelegate")
	KeyRevokePeriod = []byte("RevokePeriod")
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
	Percentage   Percentage `json:"percentage" yaml:"percentage"`
	MinDelegate  int64      `json:"min_delegate" yaml:"min_delegate"`
	RevokePeriod int64		`json:"revoke_period" yaml:"revoke_period"`
}

// NewParams creates a new Params object
func NewParams(percentage Percentage, minDelegate int64, revokePeriod int64) Params {
	return Params{
		Percentage:   percentage,
		MinDelegate:  minDelegate,
		RevokePeriod: revokePeriod,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyPercentage, &p.Percentage, validatePercentage),
		params.NewParamSetPair(KeyMinDelegate, &p.MinDelegate, validateMinDelegate),
		params.NewParamSetPair(KeyRevokePeriod, &p.RevokePeriod, validateRevokePeriod),
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
		DefaultMinDelegate,
		DefaultRevokePeriod,
	)
}

func (p Params) Validate() error {
	if err := validatePercentage(p.Percentage); err != nil {
		return errors.Wrap(err, "invalid Percentage")
	}
	if err := validateMinDelegate(p.MinDelegate); err != nil {
		return errors.Wrap(err, "invalid MinDelegate")
	}
	if err := validateRevokePeriod(p.RevokePeriod); err != nil { return errors.Wrap(err, "invalid RevokePeriod") }
	return nil
}

func validatePercentage(i interface{}) error {
	p, ok := i.(Percentage)
	if !ok {
		return errors.Errorf("invalid Percentage parameter type: %T", i)
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

func validateMinDelegate(i interface{}) error {
	md, ok := i.(int64)
	if !ok {
		return errors.Errorf("invalid MinDelegate parameter type: %T", i)
	}
	if md < 1 {
		return errors.New("minimal delegation must be at least 1")
	}
	return nil
}

func validateRevokePeriod(i interface{}) error {
	rp, ok := i.(int64)
	if !ok {
		return errors.Errorf("invalid RevokePeriod parameter type: %T", i)
	}
	if rp < 1 {
		return errors.New("RevokePeriod must be at least 1")
	}
	return nil
}
