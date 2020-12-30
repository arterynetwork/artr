package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter store keys
var (
	KeyParamSigners = []byte("Signers")
)

// ParamKeyTable for vpn module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for vpn at genesis
type Params struct {
	Signers []sdk.AccAddress `json:"signers" yaml:"signers"`
}

// NewParams creates a new Params object
func NewParams() Params {
	return Params{}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf(`
Signers: %v
	`, p.Signers)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyParamSigners, &p.Signers, validateSigners),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams()
}

func (p Params) Validate() error {
	if err := validateSigners(p.Signers); err != nil {
		return err
	}
	return nil
}

func validateSigners(i interface{}) error {
	val, ok := i.([]sdk.AccAddress)
	if !ok {
		return fmt.Errorf("invalid VPN traffic signers parameter type: %T", i)
	}

	if len(val) == 0 {
		return fmt.Errorf("empty VPN traffic signer list")
	}
	for i, s := range val {
		if s.Empty() {
			return fmt.Errorf("empty VPN traffic signer address (#%d)", i)
		}
	}

	return nil
}
