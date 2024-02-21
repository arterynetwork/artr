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

func (p Percentage) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (p Percentage) Validate() error {
	if p.Minimal <= 0 {
		return errors.New("minimal percent is non-positive")
	}
	if p.ThousandPlus <= 0 {
		return errors.New("1000+ percent is non-positive")
	}
	if p.TenKPlus <= 0 {
		return errors.New("10k+ percent is non-positive")
	}
	if p.HundredKPlus <= 0 {
		return errors.New("100k+ percent is non-positive")
	}
	if p.Minimal > p.ThousandPlus {
		return errors.New("minimal percent is reater than 1000+ one")
	}
	if p.ThousandPlus > p.TenKPlus {
		return errors.New("1000+ percent is greater than 10k+ one")
	}
	if p.TenKPlus > p.HundredKPlus {
		return errors.New("10k+ percent is greater than 100k+ one")
	}
	return nil
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

func ValidatePercentageRanges(ladder []PercentageRange) error {
	if len(ladder) == 0 {
		return errors.New("at least one range is required")
	}
	var prevStep PercentageRange
	for index, step := range ladder {
		if err := step.Validate(); err != nil {
			return errors.Wrapf(err, "invalid PercentageRange #%d", index)
		}
		if index != 0 {
			if step.Start <= prevStep.Start {
				return errors.Errorf("range #%d start (%d) less or equal than range #%d start (%d)", index+1, step.Start, index, prevStep.Start)
			}
			if step.Percent.LT(prevStep.Percent) {
				return errors.Errorf("range #%d percent (%s) less than range #%d percent (%s)", index+1, step.Percent, index, prevStep.Percent)
			}
		}
		prevStep = step
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

func ValidatePercentageTable(ladder []PercentageListRange) error {
	if len(ladder) == 0 {
		return errors.New("at least one range is required")
	}
	var prevStep PercentageListRange
	for index, step := range ladder {
		if err := step.Validate(); err != nil {
			return errors.Wrapf(err, "invalid PercentageRange #%d", index)
		}
		if index == 0 {
			if step.Start != 0 {
				return errors.Errorf("range #%d start (%d) not equal 0", index+1, step.Start)
			}
		} else {
			if step.Start <= prevStep.Start {
				return errors.Errorf("range #%d start (%d) less or equal than range #%d start (%d)", index+1, step.Start, index, prevStep.Start)
			}
		}
		prevStep = step
	}
	return nil
}
