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

func (pl PercentageListRange) String() string {
	out, _ := yaml.Marshal(pl)
	return string(out)
}

func (pl PercentageListRange) Validate() error {
	if len(pl.PercentList) != 5 {
		return errors.New("number of percent in list is not equal to 5")
	}
	for i, p := range pl.PercentList {
		if p.IsNegative() {
			return errors.Errorf("percent #%d is negative", i)
		}
	}
	return nil
}
