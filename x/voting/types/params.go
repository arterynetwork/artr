package types

import (
	"github.com/arterynetwork/artr/util"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
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

// Params - used for initializing default parameter for voting at genesis
type Params struct {
	VotingPeriod int32 `json:"voting_period" yaml:"voting_period"`
}

// NewParams creates a new Params object
func NewParams(votingPeriod int32) Params {
	return Params{
		VotingPeriod: votingPeriod,
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf(`
		VotingPeriod: #{p.VotingPeriod}
	`)
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
		return err
	}
	return nil
}

func validateVotingPeriod(i interface{}) error {
	v, ok := i.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 100 {
		return fmt.Errorf("validating period must be more then 100 blocks: %d", v)
	}

	return nil
}
