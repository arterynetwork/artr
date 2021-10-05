package types

import (
	"fmt"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

	DefaultVotingPeriod int32 = util.BlocksOneDay
)

// Parameter store keys
var (
	KeyParamVotingPeriod = []byte("VotingPeriod")
)

// ParamKeyTable for voting module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(votingPeriod int32) Params {
	return Params{
		VotingPeriod: votingPeriod,
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
		params.NewParamSetPair(KeyParamVotingPeriod, &p.VotingPeriod, validateVotingPeriod),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(DefaultVotingPeriod)
}

func (p Params) Validate() error {
	if err := validateVotingPeriod(p.VotingPeriod); err != nil {
		return errors.Wrap(err, "invalid voting_period")
	}
	return nil
}

func validateVotingPeriod(i interface{}) error {
	v, ok := i.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("validating period must be at least 1 hour: %d", v)
	}

	return nil
}
