package types

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func NewRecord() Record {
	return Record{}
}

func (x Record) IsEmpty() bool {
	return x.Requests == nil && x.NextAccrue == nil
}

func (p PercentageRange) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (p PercentageRange) Validate() error {
	if p.Percent.IsNegative() {
		return errors.New("percent is negative")
	}
	return nil
}
