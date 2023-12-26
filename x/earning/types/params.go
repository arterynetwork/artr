package types

import (
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter store keys
var (
	// default paramspace keys
	KeySigners = []byte("Signers")
)

// ParamKeyTable for earning module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(signers []sdk.AccAddress) *Params {
	res := &Params{
		Signers: make([]string, len(signers)),
	}
	for i, acc := range signers {
		res.Signers[i] = acc.String()
	}
	return res
}

func (p Params) GetSigners() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(p.Signers))
	for i, bech32 := range p.Signers {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func (p Params) String() string {
	str, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(str)
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySigners, &p.Signers, validateSigners),
	}
}

func (p Params) Validate() error {
	if err := validateSigners(p.Signers); err != nil {
		return err
	}
	return nil
}

func validateSigners(value interface{}) error {
	accz, ok := value.([]string)
	if !ok {
		return fmt.Errorf("unexpected Signers type: %T", value)
	}
	if len(accz) == 0 {
		return fmt.Errorf("signers list is empty")
	}
	for i, acc := range accz {
		if _, err := sdk.AccAddressFromBech32(acc); err != nil {
			return errors.Wrapf(err, "invalid signer address (#%d)", i)
		}
	}
	return nil
}
