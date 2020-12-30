package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter store keys
var (
	KeyInitialHeight = []byte("InitialHeight")
)

// ParamKeyTable for schedule module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for schedule at genesis
type Params struct {
	InitialHeight int64 `json:"initial_height"`
}

// NewParams creates a new Params object
func NewParams() Params {
	return Params{}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf(`
InitialHeight: %d
	`, p.InitialHeight)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyInitialHeight, &p.InitialHeight, validateInitialHeight),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams()
}

func validateInitialHeight(i interface{}) error {
	val, ok := i.(int64)
	if !ok {
		return fmt.Errorf("unexpected InitialHeight type: %T", i)
	}
	if val < 0 {
		return fmt.Errorf("initial height must be non-negative")
	}
	return nil
}
