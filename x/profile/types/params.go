package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

	DefaultFee       int64  = 1000000
	DefaultCardMagic uint64 = 0x1A21A4B61B
)

// Parameter store keys
var (
	DefaultCreators  []sdk.AccAddress = nil

	KeyCreators     = []byte("Creators")
	KeyFee          = []byte("Fee")
	KeyCardMagic    = []byte("CardMagic")
)

// ParamKeyTable for profile module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for profile at genesis
type Params struct {
	Creators  []sdk.AccAddress `json:"creators" yaml:"creators"`
	Fee       int64            `json:"fee" yaml:"fee"`
	CardMagic uint64           `json:"card_magic" yaml:"card_magic"`
}

// NewParams creates a new Params object
func NewParams(creators []sdk.AccAddress, fee int64, cardMagic uint64) Params {
	return Params{
		Creators:  creators,
		Fee:       fee,
		CardMagic: cardMagic,
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf(`
Creators: %s
Fee: %d
CardMagic: %d
`, p.Creators, p.Fee, p.CardMagic)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyCreators, &p.Creators, validateCreators),
		params.NewParamSetPair(KeyFee, &p.Fee, validateFee),
		params.NewParamSetPair(KeyCardMagic, &p.CardMagic, validateCardMagic),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultCreators,
		DefaultFee,
		DefaultCardMagic,
	)
}

func (p Params) Validate() error {
	if err := validateCreators(p.Creators); err != nil { return err }
	if err := validateFee(p.Fee); err != nil { return err }
	if err := validateCardMagic(p.CardMagic); err != nil { return err }

	return nil
}

func validateCreators(i interface{}) error {
	v, ok := i.([]sdk.AccAddress)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return fmt.Errorf("invalid creators list len: %d", len(v))
	}

	return nil
}

func validateFee(i interface{}) error {
	_, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateCardMagic(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid CardMagic parameter type: %T", i)
	}

	return nil
}
