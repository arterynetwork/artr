package types

import (
	"github.com/arterynetwork/artr/util"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace    = ModuleName
	DefaultMaxValidators = 100
	DefaultJailAfter     = 2
	DefaultUnjailAfter   = util.BlocksOneHour
)

// Parameter store keys
var (
	KeyMaxValidators = []byte("MaxValidators")
	KeyJailAfter     = []byte("JailAfter")
	KeyUnjailAfter   = []byte("UnjailAfter")
)

// ParamKeyTable for noding module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for noding at genesis
type Params struct {
	// MaxValidators - maximum count of validators that can be chosen for tendermint consensus
	MaxValidators uint16 `json:"max_validators"`
	// JailAfter - number of missed in row blocks after which a validator is jailed
	JailAfter     uint16 `json:"jail_after"`
	// UnjailAfter - number of block after which a jailed validator may unjail
	UnjailAfter   int64  `json:"unjail_after"`
}

// NewParams creates a new Params object
func NewParams(maxValidators uint16, jailAfter uint16, unjailAfter int64) Params {
	return Params{
		MaxValidators: maxValidators,
		JailAfter:     jailAfter,
		UnjailAfter:   unjailAfter,
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf(`MaxValidators: %d; JailAfter: %d; UnjailAfter: %d`,
		p.MaxValidators, p.JailAfter, p.UnjailAfter,
	)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
		params.NewParamSetPair(KeyJailAfter, &p.JailAfter, validateJailAfter),
		params.NewParamSetPair(KeyUnjailAfter, &p.UnjailAfter, validateUnjailAfter),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(DefaultMaxValidators, DefaultJailAfter, DefaultUnjailAfter)
}

func validateMaxValidators(value interface{}) error {
	x, ok := value.(uint16)
	if !ok { return fmt.Errorf("invalid parameter type: %T", value) }
	if x == 0 { return fmt.Errorf("max validators must be positive: %d", x) }
	return nil
}

func validateJailAfter(value interface{}) error {
	x, ok := value.(uint16)
	if !ok { return fmt.Errorf("invalid parameter type: %T", value) }
	if x == 0 { return fmt.Errorf("jail after must be positive: %d", x) }
	return nil
}

func validateUnjailAfter(value interface{}) error {
	x, ok := value.(int64)
	if !ok { return fmt.Errorf("invalid parameter type: %T", value) }
	if x <= 0 { return fmt.Errorf("ujail after must be positive: %d", x) }
	return nil
}

func (p *Params) Validate() error {
	if p == nil { return fmt.Errorf("params are nil") }
	if err := validateMaxValidators(p.MaxValidators); err != nil {
		return sdkerrors.Wrap(err, "invalid MaxValidators")
	}
	if err := validateJailAfter(p.JailAfter); err != nil {
		return sdkerrors.Wrap(err, "invalid JailAfter")
	}
	if err := validateUnjailAfter(p.UnjailAfter); err != nil {
		return sdkerrors.Wrap(err, "invalid UnjailAfter")
	}
	return nil
}