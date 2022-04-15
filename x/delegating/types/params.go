package types

import (
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

	DefaultMinimalPercent      = 21
	DefaultThousandPlusPercent = 24
	DefaultTenKPlusPercent     = 27
	DefaultHundredKPlusPercent = 30

	DefaultMinDelegate  = 1000
	DefaultRevokePeriod = 14
)
var (
	DefaultValidatorBonus = util.Percent(0)
)

// Parameter store keys
var (
	KeyPercentage     = []byte("Percentage")
	KeyMinDelegate    = []byte("MinDelegate")
	KeyRevokePeriod   = []byte("RevokePeriod")
	KeyValidatorBonus = []byte("ValidatorBonus")
)

// ParamKeyTable for delegating module
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p Percentage) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func NewPercentage(minimal int, oneK int, tenK int, hundredK int) *Percentage {
	return &Percentage{
		Minimal:      int64(minimal),
		ThousandPlus: int64(oneK),
		TenKPlus:     int64(tenK),
		HundredKPlus: int64(hundredK),
	}
}

func (p Percentage) Validate() error { return validatePercentage(p) }

// NewParams creates a new Params object
func NewParams(percentage Percentage, minDelegate int64, revokePeriod uint32, validatorBonus util.Fraction) *Params {
	return &Params{
		Percentage:     percentage,
		MinDelegate:    minDelegate,
		RevokePeriod:   revokePeriod,
		ValidatorBonus: validatorBonus,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(KeyPercentage, &p.Percentage, validatePercentage),
		paramTypes.NewParamSetPair(KeyMinDelegate, &p.MinDelegate, validateMinDelegate),
		paramTypes.NewParamSetPair(KeyRevokePeriod, &p.RevokePeriod, validateRevokePeriod),
		paramTypes.NewParamSetPair(KeyValidatorBonus, &p.ValidatorBonus, validateValidatorBonus),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() *Params {
	return NewParams(
		*NewPercentage(
			DefaultMinimalPercent,
			DefaultThousandPlusPercent,
			DefaultTenKPlusPercent,
			DefaultHundredKPlusPercent,
		),
		DefaultMinDelegate,
		DefaultRevokePeriod,
		DefaultValidatorBonus,
	)
}

func (p Params) Validate() error {
	if err := validatePercentage(p.Percentage); err != nil {
		return errors.Wrap(err, "invalid Percentage")
	}
	if err := validateMinDelegate(p.MinDelegate); err != nil {
		return errors.Wrap(err, "invalid MinDelegate")
	}
	if err := validateRevokePeriod(p.RevokePeriod); err != nil {
		return errors.Wrap(err, "invalid RevokePeriod")
	}
	if err := validateValidatorBonus(p.ValidatorBonus); err != nil {
		return errors.Wrap(err, "invalid ValidatorBonus")
	}
	return nil
}

func (p Params) String() string {
	bz, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

func (p Params) GetRevokePeriod(sk ScheduleKeeper, ctx sdk.Context) time.Duration {
	return time.Duration(p.RevokePeriod) * sk.OneDay(ctx)
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
	rp, ok := i.(uint32)
	if !ok {
		return errors.Errorf("invalid RevokePeriod parameter type: %T", i)
	}
	if rp < 1 {
		return errors.New("RevokePeriod must be at least 1")
	}
	return nil
}

func validateValidatorBonus(i interface{}) error {
	vb, ok := i.(util.Fraction)
	if !ok { return errors.Errorf("invalid ValidatorBonus parameter type: %T", i) }
	if vb.IsNullValue() { return errors.New("ValidatorBonus must be non-null") }
	if vb.IsNegative() { return errors.New("ValidatorBonus must be non-negative") }
	return nil
}
