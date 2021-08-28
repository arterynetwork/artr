package types

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
)

var (
	ParamStoreKeyMinSend        = []byte("minsend")
	ParamStoreKeyDustDelegation = []byte("dustd")
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(ParamStoreKeyMinSend, &p.MinSend, validateMinSend),
		paramTypes.NewParamSetPair(ParamStoreKeyDustDelegation, &p.DustDelegation, validateDustDelegation),
	}
}

// NewParams creates a new parameter configuration for the bank module
func NewParams(minSend int64) Params {
	return Params{
		MinSend:        minSend,
		DustDelegation: 0,
	}
}

// Validate all bank module parameters
func (p Params) Validate() error {
	if err := validateMinSend(p.MinSend); err != nil {
		return errors.Wrap(err, "invalid min_send")
	}
	if err := validateDustDelegation(p.DustDelegation); err != nil {
		return errors.Wrap(err, "invalid dust_delegation")
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMinSend(i interface{}) error {
	_, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDustDelegation(i interface{}) error {
	dt, ok := i.(int64)
	if !ok {
		return errors.Errorf("invalid DustDelegation parameter type: %T", i)
	}
	if dt < 0 {
		return errors.New("DustDelegation must be non-negative")
	}
	return nil
}
