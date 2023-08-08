package types

import (
	"time"

	"github.com/pkg/errors"

	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyDayNanos = []byte("DayNanos")
)

const (
	DefaultDayNanos = uint64(24 * time.Hour)
)

func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParameters() Params {
	return Params{
		DayNanos: DefaultDayNanos,
	}
}

func (p *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(KeyDayNanos, &p.DayNanos, validateDayNanos),
	}
}

func (p Params) Validate() error {
	if err := validateDayNanos(p.DayNanos); err != nil {
		return err
	}
	return nil
}

func validateDayNanos(i interface{}) error {
	wrap := func(err error) error { return errors.Wrap(err, "invalid schedule.day_nanos") }
	nanos, ok := i.(uint64)
	if !ok {
		return wrap(errors.Errorf("invalid type: %T", i))
	}
	if nanos == 0 {
		return wrap(errors.New("value must be positive"))
	}
	if nanos > DefaultDayNanos {
		return wrap(errors.Errorf("time quotient is decelarating (%d > %d)", nanos, DefaultDayNanos))
	}
	if min := uint64(time.Minute); nanos < min {
		return wrap(errors.Errorf("too fast (%d < %d)", nanos, min))
	}
	return nil
}
