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

func TestCase_MarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		c       Case
		want    []byte
		wantErr bool
	}{
		{name: "shouldMarshallToJsonLower", c: Lower, want: []byte("\"lower\""), wantErr: false},
		{name: "shouldMarshallToJsonUpper", c: Upper, want: []byte("\"upper\""), wantErr: false},
		{name: "shouldMarshallToJsonNone", c: None, want: []byte("\"none\""), wantErr: false},
		{name: "shouldMarshallToJsonCapitalize", c: Capitalize, want: []byte("\"capitalize\""), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCase_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		c       Case
		args    args
		want    Case
		wantErr bool
	}{
		{name: "shouldUnmarshallToJsonLower", args: args{b: []byte("\"lower\"")}, c: *new(Case), want: Lower, wantErr: false},
		{name: "shouldUnmarshallToJsonUpper", args: args{b: []byte("\"upper\"")}, c: *new(Case), want: Upper, wantErr: false},
		{name: "shouldUnmarshallToJsonNone", args: args{b: []byte("\"none\"")}, c: *new(Case), want: None, wantErr: false},
		{name: "shouldUnmarshallToJsonCapitalize", args: args{b: []byte("\"capitalize\"")}, c: *new(Case), want: Capitalize, wantErr: false},
		{name: "shouldNotUnmarshallUnknownCase", args: args{b: []byte("\"noop\"")}, c: *new(Case), want: *new(Case), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.c, tt.want) {
				t.Errorf("UnmarshalJSON() got = %v, want %v", tt.c, tt.want)
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
