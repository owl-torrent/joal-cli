package casing

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type Case int

const (
	None Case = iota
	Upper
	Lower
	Capitalize
)

func (c Case) String() string {
	return toString[c]
}

func (c Case) ApplyCase(str string) string {
	switch c {
	case None:
		return str
	case Lower:
		return strings.ToLower(str)
	case Upper:
		return strings.ToUpper(str)
	case Capitalize:
		return strings.Title(str)
	}
	return str
}

var toString = map[Case]string{
	None:       "none",
	Lower:      "lower",
	Upper:      "upper",
	Capitalize: "capitalize",
}

var toID = map[string]Case{
	"none":       None,
	"lower":      Lower,
	"upper":      Upper,
	"capitalize": Capitalize,
}

func (c Case) MarshalYAML() ([]byte, error) {
	return []byte(toString[c]), nil
}

func (c *Case) UnmarshalYAML(value *yaml.Node) error {
	var j string
	err := value.Decode(&j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Lower' in this case.
	cc, ok := toID[j]
	if !ok {
		return fmt.Errorf("failed to unmarshall Unknown Case '%s'", j)
	}
	*c = cc
	return nil
}
