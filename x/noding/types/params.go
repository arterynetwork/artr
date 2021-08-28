package types

import (
	"fmt"
	"gopkg.in/yaml.v2"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
)

// Default parameter namespace
const (
	DefaultParamspace        = ModuleName
	DefaultMaxValidators     = 100
	DefaultJailAfter         = 2
	DefaultUnjailAfter       = util.BlocksOneHour
	DefaultLotteryValidators = 0
	DefaultMinStatus         = referral.StatusLeader
)

// Parameter store keys
var (
	KeyMaxValidators     = []byte("MaxValidators")
	KeyJailAfter         = []byte("JailAfter")
	KeyUnjailAfter       = []byte("UnjailAfter")
	KeyLotteryValidators = []byte("LotteryValidators")
	KeyMinStatus         = []byte("MinStatus")
)

// ParamKeyTable for noding module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(maxValidators, jailAfter, unjailAfter, lotteryValidators uint32, minStatus referral.Status) Params {
	return Params{
		MaxValidators:     maxValidators,
		JailAfter:         jailAfter,
		UnjailAfter:       unjailAfter,
		LotteryValidators: lotteryValidators,
		MinStatus:         minStatus,
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	bz, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
		params.NewParamSetPair(KeyJailAfter, &p.JailAfter, validateJailAfter),
		params.NewParamSetPair(KeyUnjailAfter, &p.UnjailAfter, validateUnjailAfter),
		params.NewParamSetPair(KeyLotteryValidators, &p.LotteryValidators, validateAdditionalValidators),
		params.NewParamSetPair(KeyMinStatus, &p.MinStatus, validateStatus),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxValidators,
		DefaultJailAfter,
		DefaultUnjailAfter,
		DefaultLotteryValidators,
		DefaultMinStatus,
	)
}

func validateMaxValidators(value interface{}) error {
	x, ok := value.(uint32)
	if !ok {
		return fmt.Errorf("invalid max_validators type: %T", value)
	}
	if x == 0 {
		return fmt.Errorf("max_validators must be positive: %d", x)
	}
	return nil
}

func validateAdditionalValidators(value interface{}) error {
	_, ok := value.(uint32)
	if !ok {
		return fmt.Errorf("invalid lottery_validators type: %T", value)
	}
	return nil
}

func validateJailAfter(value interface{}) error {
	x, ok := value.(uint32)
	if !ok {
		return fmt.Errorf("invalid jail_after type: %T", value)
	}
	if x == 0 {
		return fmt.Errorf("jail after must be positive: %d", x)
	}
	return nil
}

func validateUnjailAfter(value interface{}) error {
	x, ok := value.(uint32)
	if !ok {
		return fmt.Errorf("invalid unjail_after type: %T", value)
	}
	if x <= 0 {
		return fmt.Errorf("ujail after must be positive: %d", x)
	}
	return nil
}

func validateStatus(i interface{}) error {
	status, ok := i.(referral.Status)
	if !ok {
		return fmt.Errorf("invalid min_status type (uint8 expected): %T", i)
	}
	if status < referral.StatusLucky {
		return fmt.Errorf("status beyond min: %d", status)
	}
	if status > referral.StatusAbsoluteChampion {
		return fmt.Errorf("status above max: %d", status)
	}
	return nil
}

func (p *Params) Validate() error {
	if p == nil {
		return fmt.Errorf("params are nil")
	}
	if err := validateMaxValidators(p.MaxValidators); err != nil {
		return sdkerrors.Wrap(err, "invalid MaxValidators")
	}
	if err := validateJailAfter(p.JailAfter); err != nil {
		return sdkerrors.Wrap(err, "invalid JailAfter")
	}
	if err := validateUnjailAfter(p.UnjailAfter); err != nil {
		return sdkerrors.Wrap(err, "invalid UnjailAfter")
	}
	if err := validateAdditionalValidators(p.LotteryValidators); err != nil {
		return sdkerrors.Wrap(err, "invalid LotteryValidators")
	}
	if err := validateStatus(p.MinStatus); err != nil {
		return sdkerrors.Wrap(err, "invalid MinStatus")
	}
	return nil
}
