package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/casing"
	"github.com/anthonyraymond/joal-cli/pkg/randutils"
	"github.com/pkg/errors"
	"strings"
)

type HashAlgorithm struct {
	TrimLeadingZeroes bool        `yaml:"trimLeadingZeroes"`
	MaxLength         int         `yaml:"maxLength"`
	Case              casing.Case `yaml:"case"`
}

func (a *HashAlgorithm) Generate() string {
	const chars = "0123456789ABCDEF"
	hex := randutils.String(chars, a.MaxLength)
	if a.TrimLeadingZeroes {
		hex = strings.TrimLeft(hex, "0")
	}
	if len(hex) == 0 { // if there was only zeroes and we trimmed them try again
		return a.Generate()
	}
	return a.Case.ApplyCase(hex)
}

func (a *HashAlgorithm) AfterPropertiesSet() error {
	if a.MaxLength < 1 {
		return errors.New("property 'MaxLength' must be greater than 1 in NumRangeAsHexadecimalAlgorithm ")
	}

	return nil
}
