package util

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

type Fraction struct {
	num   int64
	denom int64
}

func (x Fraction) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}

func (x *Fraction) UnmarshalJSON(bz []byte) error {
	var s string
	err := json.Unmarshal(bz, &s)
	if err != nil { return err }
	y, err := ParseFraction(s)
	if err != nil { return err }

	x.num = y.num
	x.denom = y.denom
	return nil
}

func (x Fraction) MarshalAmino() (string, error) { return x.String(), nil }

func (x *Fraction) UnmarshalAmino(s string) error {
	y, err := ParseFraction(s)
	if err != nil { return err }

	x.num = y.num
	x.denom = y.denom
	return nil
}

func (x Fraction) MarshalYAML() (string, error) { return x.String(), nil }

func NewFraction(num int64, denom int64) Fraction {
	return Fraction{
		num:   num,
		denom: denom,
	}
}

func Percent(x int64) Fraction {
	return NewFraction(x, 100)
}

func Permille(x int64) Fraction {
	return NewFraction(x, 1000)
}

func FractionFromInt64(x int64) Fraction { return NewFraction(x, 1) }

func FractionZero() Fraction { return FractionFromInt64(0) }

func ParseFraction(s string) (Fraction, error) {
	var (
		num, denom int64
		err error
	)

	// Try percent notation first
	pattern := regexp.MustCompile(`^(-?\d+)%$`)
	match := pattern.FindStringSubmatch(s)
	if len(match) > 0 {
		denom = 100
	} else {
		pattern = regexp.MustCompile(`^(-?\d+)(?:/(\d+))?$`)
		match = pattern.FindStringSubmatch(s)
		if len(match) == 0 {
			return Fraction{}, fmt.Errorf("invalid fraction format, 'x%' or 'x/y' expected: " + s)
		}
		if len(match) > 2 {
			denom, err = strconv.ParseInt(match[2], 10, 64)
			if err != nil { return Fraction{}, err }
		} else {
			denom = 1
		}
	}

	num, err = strconv.ParseInt(match[1], 10, 64)
	if err != nil { return Fraction{}, err }

	return NewFraction(num, denom), nil
}

func (x Fraction) String() string {
	if x.denom == 100 {
		return fmt.Sprintf("%d%%", x.num)
	}
	return fmt.Sprintf("%d/%d", x.num, x.denom)
}

func (x Fraction) Float64() float64 {
	return float64(x.num) / float64(x.denom)
}

func (x Fraction) Int64() int64 {
	return x.num / x.denom
}

func (x Fraction) Reduce() Fraction {
	if x.denom < 0 {
		x = NewFraction(-x.num, -x.denom)
	}
	q := gcd(x.num, x.denom)
	return NewFraction(x.num / q, x.denom / q)
}

func (x Fraction) Mul(y Fraction) Fraction {
	return NewFraction(x.num * y.num, x.denom * y.denom).Reduce()
}

func (x Fraction) MulInt64(y int64) Fraction {
	return x.Mul(NewFraction(y, 1))
}

func (x Fraction) Div(y Fraction) Fraction {
	return NewFraction(x.num * y.denom, x.denom * y.num).Reduce()
}

func (x Fraction) DivInt64(y int64) Fraction {
	return x.Div(NewFraction(y, 1))
}

func (x Fraction) Add(y Fraction) Fraction {
	denom := lcm(x.denom, y.denom)
	return NewFraction(x.num * denom/x.denom + y.num * denom/y.denom, denom)
}

func (x Fraction) Sub(y Fraction) Fraction {
	denom := lcm(x.denom, y.denom)
	return NewFraction(x.num * denom/x.denom - y.num * denom/y.denom, denom)
}

func (x Fraction) IsZero() bool {
	return x.num == 0
}

func (x Fraction) IsNegative() bool {
	return x.num < 0 && x.denom > 0 || x.num > 0 && x.denom < 0
}

func (x Fraction) IsPositive() bool {
	return x.num > 0 && x.denom > 0 || x.num < 0 && x.denom < 0
}

func (x Fraction) GT(y Fraction) bool { return x.Sub(y).IsPositive() }
func (x Fraction) LT(y Fraction) bool { return y.GT(x) }
func (x Fraction) GTE(y Fraction) bool { return !x.LT(y) }
func (x Fraction) LTE(y Fraction) bool { return !x.GT(y) }

func gcd(x int64, y int64) int64 {
	if x < 0 {
		x = -x
	}
	if y < 0 {
		y = -y
	}
	if x == 0 {
		return y
	}
	if y == 0 {
		return x
	}
	for {
		q := x / y
		r := x - q * y
		if r == 0 {
			return y
		}
		x, y = y, r
	}
}

func lcm(x int64, y int64) int64 {
	r := x * y / gcd(x, y)
	if r < 0 {
		r = -r
	}
	return r
}

