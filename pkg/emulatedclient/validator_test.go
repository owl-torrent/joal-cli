package emulatedclient

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_TagNameFunction_ShouldMatchName(t *testing.T) {
	s := struct {
		JsonName  string `json:"naming"`
		JsonName2 string `json:"abcd,omitempty"`
		YamlName  string `yaml:"defg"`
		YamlName2 string `yaml:"ghij,omitempty"`
		Empty     string
	}{}

	assert.Equal(t, "naming", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("JsonName"))))
	assert.Equal(t, "abcd", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("JsonName2"))))
	assert.Equal(t, "defg", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("YamlName"))))
	assert.Equal(t, "ghij", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("YamlName2"))))
	assert.Equal(t, "", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("Empty"))))
}

//noinspection GoUnusedParameter
func mustFieldByName(st reflect.StructField, b bool) reflect.StructField {
	return st
}
