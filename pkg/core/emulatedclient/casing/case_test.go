package casing

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"reflect"
	"testing"
)

func TestCase_MarshalYaml(t *testing.T) {
	tests := []struct {
		name    string
		c       Case
		want    []byte
		wantErr bool
	}{
		{name: "shouldMarshallToYamlLower", c: Lower, want: []byte("lower"), wantErr: false},
		{name: "shouldMarshallToYamlUpper", c: Upper, want: []byte("upper"), wantErr: false},
		{name: "shouldMarshallToYamlNone", c: None, want: []byte("none"), wantErr: false},
		{name: "shouldMarshallToYamlCapitalize", c: Capitalize, want: []byte("capitalize"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.MarshalYAML()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalYAML() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCase_UnmarshalYamlLower(t *testing.T) {
	type tmp struct {
		Case *Case `yaml:"case"`
	}
	yamlInput := "case: lower"

	res := new(tmp)
	err := yaml.Unmarshal([]byte(yamlInput), res)
	if err != nil {
		t.Errorf("Failed to unmarshall yaml '%s': %v", yamlInput, err)
	}
	assert.Equal(t, Lower, *res.Case)
}

func TestCase_UnmarshalYamlUpper(t *testing.T) {
	type tmp struct {
		Case *Case `yaml:"case"`
	}
	yamlInput := "case: upper"

	res := new(tmp)
	err := yaml.Unmarshal([]byte(yamlInput), res)
	if err != nil {
		t.Errorf("Failed to unmarshall yaml '%s': %v", yamlInput, err)
	}
	assert.Equal(t, Upper, *res.Case)
}

func TestCase_UnmarshalYamlCapitalize(t *testing.T) {
	type tmp struct {
		Case *Case `yaml:"case"`
	}
	yamlInput := "case: capitalize"

	res := new(tmp)
	err := yaml.Unmarshal([]byte(yamlInput), res)
	if err != nil {
		t.Errorf("Failed to unmarshall yaml '%s': %v", yamlInput, err)
	}
	assert.Equal(t, Capitalize, *res.Case)
}

func TestCase_UnmarshalYamlNone(t *testing.T) {
	type tmp struct {
		Case *Case `yaml:"case"`
	}
	yamlInput := "case: none"

	res := new(tmp)
	err := yaml.Unmarshal([]byte(yamlInput), res)
	if err != nil {
		t.Errorf("Failed to unmarshall yaml '%s': %v", yamlInput, err)
	}
	assert.Equal(t, None, *res.Case)
}

func TestCase_UnmarshalYamlUnknown(t *testing.T) {
	type tmp struct {
		Case *Case `yaml:"case"`
	}
	yamlInput := "case: noop"

	res := new(tmp)
	err := yaml.Unmarshal([]byte(yamlInput), res)
	if err == nil {
		t.Error("Unmarshall should have failed")
	}
}

func TestCase_ApplyCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		c    Case
		args args
		want string
	}{
		{name: "shouldApplyNone", c: None, args: args{str: "AbCd"}, want: "AbCd"},
		{name: "shouldApplyLower", c: Lower, args: args{str: "AbCd"}, want: "abcd"},
		{name: "shouldApplyUpper", c: Upper, args: args{str: "AbCd"}, want: "ABCD"},
		{name: "shouldApplyCapitalize", c: Capitalize, args: args{str: "abcd"}, want: "Abcd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.ApplyCase(tt.args.str); got != tt.want {
				t.Errorf("ApplyCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
