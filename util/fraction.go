package util

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"math/big"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type Fraction struct {
	num   *big.Int
	denom *big.Int
}

var (
	_ json.Marshaler   = new(Fraction)
	_ yaml.Marshaler   = new(Fraction)
	_ json.Unmarshaler = new(Fraction)
	_ yaml.Unmarshaler = new(Fraction)
)

func (x Fraction) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}

func (x *Fraction) UnmarshalJSON(bz []byte) error {
	var s string
	err := json.Unmarshal(bz, &s)
	if err != nil {
		return err
	}
	return x.unmarshalText(s)
}

func (x Fraction) MarshalAmino() (string, error) { return x.String(), nil }

func (x *Fraction) UnmarshalAmino(s string) error {
	return x.unmarshalText(s)
}

func (x Fraction) MarshalYAML() (interface{}, error) {
	return x.String(), nil
}

func (x *Fraction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return x.unmarshalText(s)
}

func (x *Fraction) unmarshalText(s string) error {
	y, err := ParseFraction(s)
	if err != nil {
		return err
	}

	x.num = y.num
	x.denom = y.denom
	return nil
}

func NewFraction(num int64, denom int64) Fraction {
	return Fraction{
		num:   big.NewInt(num),
		denom: big.NewInt(denom),
	}
}

func Percent(x int64) Fraction {
	return NewFraction(x, 100)
}

func Permille(x int64) Fraction {
	return NewFraction(x, 1000)
}

func FractionInt(x int64) Fraction { return NewFraction(x, 1) }

func FractionZero() Fraction { return FractionInt(0) }

func ParseFraction(s string) (Fraction, error) {
	var (
		num   = &big.Int{}
		denom = &big.Int{}
		err   error
	)

	if s == "nil" {
		return Fraction{}, nil
	}

	// Try percent notation first
	pattern := regexp.MustCompile(`^(-?\d+)%$`)
	match := pattern.FindStringSubmatch(s)
	if len(match) > 0 {
		denom = big.NewInt(100)
	} else {
		pattern = regexp.MustCompile(`^(-?\d+)(?:/(\d+))?$`)
		match = pattern.FindStringSubmatch(s)
		if len(match) == 0 {
			return Fraction{}, fmt.Errorf("invalid fraction format, 'x%' or 'x/y' expected: " + s)
		}
		if len(match) > 2 {
			err = denom.UnmarshalText([]byte(match[2]))
			if err != nil {
				return Fraction{}, err
			}
		} else {
			denom = big.NewInt(1)
		}
	}

	err = num.UnmarshalText([]byte(match[1]))
	if err != nil {
		return Fraction{}, err
	}

	return Fraction{num, denom}, nil
}

func (x Fraction) String() string {
	if x.IsNullValue() {
		return "nil"
	}
	sb := strings.Builder{}

	bytes, err := x.num.MarshalText()
	if err != nil {
		panic(errors.Wrap(err, "couldn't marshal numerator"))
	}
	sb.Write(bytes)

	if x.denom != nil && x.denom.IsInt64() && x.denom.Int64() == 100 {
		sb.WriteRune('%')
	} else {
		sb.WriteRune('/')
		bytes, err = x.denom.MarshalText()
		if err != nil {
			panic(errors.Wrap(err, "couldn't marshal denominator"))
		}
		sb.Write(bytes)
	}
	return sb.String()
}

func (x Fraction) Int64() int64 {
	return (&big.Int{}).Quo(x.num, x.denom).Int64()
}

func (x Fraction) Reduce() Fraction {
	if x.denom.Sign() < 0 {
		x.num.Neg(x.num)
		x.denom.Neg(x.denom)
	}
	q := (&big.Int{}).GCD(nil, nil, x.num, x.denom)
	x.num.Quo(x.num, q)
	x.denom.Quo(x.denom, q)
	return x
}

func (x Fraction) Mul(y Fraction) Fraction {
	return Fraction{
		(&big.Int{}).Mul(x.num, y.num),
		(&big.Int{}).Mul(x.denom, y.denom),
	}.Reduce()
}

func (x Fraction) MulInt64(y int64) Fraction {
	return x.Mul(FractionInt(y))
}

func (x Fraction) Div(y Fraction) Fraction {
	return Fraction{
		(&big.Int{}).Mul(x.num, y.denom),
		(&big.Int{}).Mul(x.denom, y.num),
	}.Reduce()
}

func (x Fraction) DivInt64(y int64) Fraction {
	return x.Div(FractionInt(y))
}

func (x Fraction) Neg() Fraction {
	return Fraction{
		(&big.Int{}).Neg(x.num),
		(&big.Int{}).Set(x.denom),
	}
}

func (x Fraction) Add(y Fraction) Fraction {
	comDenom := lcm(x.denom, y.denom)

	a := &big.Int{}
	a.Mul(x.num, comDenom)
	a.Quo(a, x.denom)
	b := &big.Int{}
	b.Mul(y.num, comDenom)
	b.Quo(b, y.denom)
	return Fraction{a.Add(a, b), comDenom}
}

func (x Fraction) Sub(y Fraction) Fraction {
	return x.Add(y.Neg())
}

func (x Fraction) IsNullValue() bool {
	return x.num == nil && x.denom == nil
}

func (x Fraction) IsZero() bool {
	return x.num.Sign() == 0
}

func (x Fraction) IsNegative() bool {
	return x.num.Sign()*x.denom.Sign() < 0
}

func (x Fraction) IsPositive() bool {
	return x.num.Sign()*x.denom.Sign() > 0
}

func (x Fraction) GT(y Fraction) bool  { return x.Sub(y).IsPositive() }
func (x Fraction) LT(y Fraction) bool  { return y.GT(x) }
func (x Fraction) GTE(y Fraction) bool { return !x.LT(y) }
func (x Fraction) LTE(y Fraction) bool { return !x.GT(y) }

func lcm(x *big.Int, y *big.Int) *big.Int {
	res := &big.Int{}
	res.Mul(x, y)
	res.Quo(res, (&big.Int{}).GCD(nil, nil, x, y))
	if res.Sign() < 0 {
		res.Neg(res)
	}
	return res
}
