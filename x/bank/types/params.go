package types

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName

	DefaultMinSend        = 1000
	DefaultDustDelegation = 0

	DefaultMaxTransactionFee = 10_000000
)

var (
	DefaultTransactionFee = util.Permille(3)
)

var (
	ParamStoreKeyMinSend           = []byte("minsend")
	ParamStoreKeyDustDelegation    = []byte("dustd")
	ParamStoreKeyTransactionFee    = []byte("txfee")
	ParamStoreKeyMaxTransactionFee = []byte("maxtxfee")
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(ParamStoreKeyMinSend, &p.MinSend, validateMinSend),
		paramTypes.NewParamSetPair(ParamStoreKeyDustDelegation, &p.DustDelegation, validateDustDelegation),
		paramTypes.NewParamSetPair(ParamStoreKeyTransactionFee, &p.TransactionFee, validateTransactionFee),
		paramTypes.NewParamSetPair(ParamStoreKeyMaxTransactionFee, &p.MaxTransactionFee, validateMaxTransactionFee),
	}
}

// NewParams creates a new parameter configuration for the bank module
func NewParams(minSend int64, dust int64, fee util.Fraction, maxFee int64) Params {
	return Params{
		MinSend:           minSend,
		DustDelegation:    dust,
		TransactionFee:    fee,
		MaxTransactionFee: maxFee,
	}
}

// DefaultParams defines the parameters for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultMinSend,
		DefaultDustDelegation,
		DefaultTransactionFee,
		DefaultMaxTransactionFee,
	)
}

// Validate all bank module parameters
func (p Params) Validate() error {
	if err := validateMinSend(p.MinSend); err != nil {
		return errors.Wrap(err, "invalid min_send")
	}
	if err := validateDustDelegation(p.DustDelegation); err != nil {
		return errors.Wrap(err, "invalid dust_delegation")
	}
	if err := validateTransactionFee(p.TransactionFee); err != nil {
		return errors.Wrap(err, "invalid transaction_fee")
	}
	if err := validateMaxTransactionFee(p.MaxTransactionFee); err != nil {
		return errors.Wrap(err, "invalid max_transaction_fee")
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

func validateTransactionFee(i interface{}) error {
	dt, ok := i.(util.Fraction)
	if !ok {
		return errors.Errorf("invalid TransactionFee parameter type: %T", i)
	}
	if dt.IsNegative() {
		return errors.New("TransactionFee must be non-negative")
	}
	return nil
}

func validateMaxTransactionFee(i interface{}) error {
	dt, ok := i.(int64)
	if !ok {
		return errors.Errorf("invalid MaxTransactionFee parameter type: %T", i)
	}
	if dt <= 0 {
		return errors.New("MaxTransactionFee must be positive")
	}
	return nil
}
