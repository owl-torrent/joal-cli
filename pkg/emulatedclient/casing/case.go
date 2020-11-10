package casing

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// MarshalJSON marshals the enum as a quoted json string
func (c Case) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[c])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (c Case) MarshalYAML() ([]byte, error) {
	return []byte(toString[c]), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (c *Case) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
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

func (c *Case) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var j string
	err := unmarshal(&j)
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
