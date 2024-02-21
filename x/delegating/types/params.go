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

	DefaultMinDelegate  = 1000
	DefaultRevokePeriod = 14
)

var (
	DefaultBurnOnRevoke          = util.Percent(5)
	DefaultAccruePercentageTable = []PercentageListRange{
		{Start: 0, PercentList: []util.Fraction{
			util.Percent(21),
			util.Percent(0),
			util.Percent(1),
			util.Percent(0),
			util.Percent(0),
		}},
		{Start: 1_000_000000, PercentList: []util.Fraction{
			util.Percent(24),
			util.Percent(0),
			util.Percent(1),
			util.Percent(0),
			util.Percent(0),
		}},
		{Start: 10_000_000000, PercentList: []util.Fraction{
			util.Percent(27),
			util.Percent(0),
			util.Percent(1),
			util.Percent(0),
			util.Percent(0),
		}},
		{Start: 100_000_000000, PercentList: []util.Fraction{
			util.Percent(30),
			util.Percent(0),
			util.Percent(1),
			util.Percent(0),
			util.Percent(0),
		}},
	}
)

// Parameter store keys
var (
	KeyMinDelegate           = []byte("MinDelegate")
	KeyRevokePeriod          = []byte("RevokePeriod")
	KeyBurnOnRevoke          = []byte("BurnOnRevoke")
	KeyAccruePercentageTable = []byte("AccruePercentageTable")
)

// ParamKeyTable for delegating module
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(minDelegate int64, revokePeriod uint32, burnOnRevoke util.Fraction, accruePercentageTable []PercentageListRange) *Params {
	return &Params{
		MinDelegate:           minDelegate,
		RevokePeriod:          revokePeriod,
		BurnOnRevoke:          burnOnRevoke,
		AccruePercentageTable: accruePercentageTable,
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() *Params {
	return NewParams(
		DefaultMinDelegate,
		DefaultRevokePeriod,
		DefaultBurnOnRevoke,
		DefaultAccruePercentageTable,
	)
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
		paramTypes.NewParamSetPair(KeyMinDelegate, &p.MinDelegate, validateMinDelegate),
		paramTypes.NewParamSetPair(KeyRevokePeriod, &p.RevokePeriod, validateRevokePeriod),
		paramTypes.NewParamSetPair(KeyBurnOnRevoke, &p.BurnOnRevoke, validateBurnOnRevoke),
		paramTypes.NewParamSetPair(KeyAccruePercentageTable, &p.AccruePercentageTable, validateAccruePercentageTable),
	}
}

func (p Params) Validate() error {
	if err := validateMinDelegate(p.MinDelegate); err != nil {
		return errors.Wrap(err, "invalid MinDelegate")
	}
	if err := validateRevokePeriod(p.RevokePeriod); err != nil {
		return errors.Wrap(err, "invalid RevokePeriod")
	}
	if err := validateBurnOnRevoke(p.BurnOnRevoke); err != nil {
		return errors.Wrap(err, "invalid BurnOnRevoke")
	}
	if err := validateAccruePercentageTable(p.AccruePercentageTable); err != nil {
		return errors.Wrap(err, "invalid AccruePercentageTable")
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

func validateAccruePercentageTable(i interface{}) error {
	v, ok := i.([]PercentageListRange)
	if !ok {
		return errors.Errorf("invalid AccruePercentageTable parameter type: %T", i)
	}
	if err := ValidatePercentageTable(v); err != nil {
		return errors.Wrap(err, "invalid AccruePercentageTable parameter:")
	}
	return nil
}

func (p Params) GetRevokePeriod(sk ScheduleKeeper, ctx sdk.Context) time.Duration {
	return time.Duration(p.RevokePeriod) * sk.OneDay(ctx)
}
