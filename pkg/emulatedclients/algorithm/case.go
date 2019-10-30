package algorithm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Case int

const (
	Lower Case = iota
	Upper
	None
)

func (c Case) String() string {
	return toString[c]
}

func (c Case) ApplyCase(str string) string {
	if c == None {
		return str
	}
	if c == Lower {
		return strings.ToLower(str)
	}
	return strings.ToUpper(str)
}

var toString = map[Case]string{
	Lower: "lower",
	Upper: "upper",
	None:  "none",
}

var toID = map[string]Case{
	"lower": Lower,
	"upper": Upper,
	"none":  None,
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
		return errors.New(fmt.Sprintf("Failed to unmarshall Unknown Case '%s'", j))
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
		return errors.New(fmt.Sprintf("Failed to unmarshall Unknown Case '%s'", j))
	}
	*c = cc
	return nil
}
