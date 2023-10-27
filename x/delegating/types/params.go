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
	DefaultValidatorBonus         = util.Percent(0)
	DefaultValidator              = util.Percent(15)
	DefaultBurnOnRevoke           = util.Percent(5)
	DefaultAccruePercentageRanges = []PercentageRange{
		{Start: 0, Percent: util.Percent(DefaultMinimalPercent)},
		{Start: 1_000_000000, Percent: util.Percent(DefaultThousandPlusPercent)},
		{Start: 10_000_000000, Percent: util.Percent(DefaultTenKPlusPercent)},
		{Start: 100_000_000000, Percent: util.Percent(DefaultHundredKPlusPercent)},
	}
)

// Parameter store keys
var (
	KeyPercentage             = []byte("Percentage")
	KeyMinDelegate            = []byte("MinDelegate")
	KeyRevokePeriod           = []byte("RevokePeriod")
	KeyValidatorBonus         = []byte("ValidatorBonus")
	KeyValidator              = []byte("Validator")
	KeyBurnOnRevoke           = []byte("BurnOnRevoke")
	KeyAccruePercentageRanges = []byte("AccruePercentageRanges")
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
func NewParams(percentage Percentage, minDelegate int64, revokePeriod uint32, validatorBonus util.Fraction, validator util.Fraction, burnOnRevoke util.Fraction, accruePercentageRanges []PercentageRange) *Params {
	return &Params{
		Percentage:             percentage,
		MinDelegate:            minDelegate,
		RevokePeriod:           revokePeriod,
		ValidatorBonus:         validatorBonus,
		Validator:              validator,
		BurnOnRevoke:           burnOnRevoke,
		AccruePercentageRanges: accruePercentageRanges,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(KeyPercentage, &p.Percentage, validatePercentage),
		paramTypes.NewParamSetPair(KeyMinDelegate, &p.MinDelegate, validateMinDelegate),
		paramTypes.NewParamSetPair(KeyRevokePeriod, &p.RevokePeriod, validateRevokePeriod),
		paramTypes.NewParamSetPair(KeyValidatorBonus, &p.ValidatorBonus, validateValidatorBonus),
		paramTypes.NewParamSetPair(KeyValidator, &p.Validator, validateValidator),
		paramTypes.NewParamSetPair(KeyBurnOnRevoke, &p.BurnOnRevoke, validateBurnOnRevoke),
		paramTypes.NewParamSetPair(KeyAccruePercentageRanges, &p.AccruePercentageRanges, validateAccruePercentageRanges),
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
		DefaultValidator,
		DefaultBurnOnRevoke,
		DefaultAccruePercentageRanges,
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
	if err := validateValidator(p.Validator); err != nil {
		return errors.Wrap(err, "invalid Validator")
	}
	if p.Validator.LT(util.Percent(p.Percentage.HundredKPlus)) {
		return errors.Errorf("Validators' percent must be greater or equal than the 100K+ one (%s < %d%%)", p.Validator, p.Percentage.HundredKPlus)
	}
	if err := validateAccruePercentageRanges(p.AccruePercentageRanges); err != nil {
		return errors.Wrap(err, "invalid AccruePercentageRanges")
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
	if !ok {
		return errors.Errorf("invalid ValidatorBonus parameter type: %T", i)
	}
	if vb.IsNullValue() {
		return errors.New("ValidatorBonus must be non-null")
	}
	if vb.IsNegative() {
		return errors.New("ValidatorBonus must be non-negative")
	}
	return nil
}

func validateValidator(i interface{}) error {
	vb, ok := i.(util.Fraction)
	if !ok {
		return errors.Errorf("invalid Validator parameter type: %T", i)
	}
	if vb.IsNullValue() {
		return errors.New("Validator must be non-null")
	}
	if vb.IsNegative() {
		return errors.New("Validator must be non-negative")
	}
	return nil
}

func validateBurnOnRevoke(i interface{}) error {
	vb, ok := i.(util.Fraction)
	if !ok {
		return errors.Errorf("invalid BurnOnRevoke parameter type: %T", i)
	}
	if vb.GT(util.Percent(100)) {
		return errors.New("BurnOnRevoke must be less than 100%")
	}
	if vb.IsNegative() {
		return errors.New("BurnOnRevoke must be non-negative")
	}
	return nil
}

func ValidatePercentageRanges(ladder []PercentageRange) error {
	if len(ladder) == 0 {
		return errors.New("at least one range is required")
	}
	var prevStep PercentageRange
	for index, step := range ladder {
		if err := step.Validate(); err != nil {
			return errors.Wrapf(err, "invalid PercentageRange #%d", index)
		}
		if index != 0 {
			if step.Start <= prevStep.Start {
				return errors.Errorf("range #%d start (%d) less or equal than range #%d start (%d)", index+1, step.Start, index, prevStep.Start)
			}
			if step.Percent.LT(prevStep.Percent) {
				return errors.Errorf("range #%d percent (%s) less than range #%d percent (%s)", index+1, step.Percent, index, prevStep.Percent)
			}
		}
		prevStep = step
	}
	return nil
}

func validateAccruePercentageRanges(i interface{}) error {
	v, ok := i.([]PercentageRange)
	if !ok {
		return errors.Errorf("invalid AccruePercentageRanges parameter type: %T", i)
	}
	if err := ValidatePercentageRanges(v); err != nil {
		return errors.Wrap(err, "invalid AccruePercentageRanges parameter:")
	}
	return nil
}
