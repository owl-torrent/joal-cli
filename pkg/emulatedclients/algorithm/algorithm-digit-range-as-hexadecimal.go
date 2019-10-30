package algorithm

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/randutils"
	"github.com/pkg/errors"
	"strconv"
)

type NumRangeAsHexadecimalAlgorithm struct {
	Min               *int64 `yaml:"min"`
	Max               *int64 `yaml:"max"`
	LeftPadWithZeroes bool   `yaml:"leftPadWithZeroes"`
	MaxLength         *int   `yaml:"maxLength"`
	Case              *Case  `yaml:"case"`
}

func (a *NumRangeAsHexadecimalAlgorithm) Generate() string {
	hex := strconv.FormatInt(randutils.Range(*a.Min, *a.Max), 16)
	if a.LeftPadWithZeroes {
		hex = fmt.Sprintf("%0"+strconv.Itoa(*a.MaxLength)+"s", hex) // left pad with zeros
	}
	if len(hex) > *a.MaxLength {
		hex = hex[len(hex)-*a.MaxLength:] // substring to keep only the 8 rightmost characters
	}
	return a.Case.ApplyCase(hex)
}

func (a *NumRangeAsHexadecimalAlgorithm) AfterPropertiesSet() error {
	if a.Min == nil {
		return errors.New("property 'min' is required in NumRangeAsHexadecimalAlgorithm")
	}
	if a.Max == nil {
		return errors.New("property 'max' is required in NumRangeAsHexadecimalAlgorithm")
	}
	if a.Min == a.Max {
		return errors.New("'max' must be greater or equal to 'min' in NumRangeAsHexadecimalAlgorithm")
	}

	return nil
}
