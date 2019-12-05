package validationutils

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_TagNameFunction_ShouldMatchName(t *testing.T) {
	s := struct {
		JsonName  string `json:"naming"`
		JsonName2 string `json:"naming,omitempty"`
		YamlName  string `yaml:"naming"`
		YamlName2 string `yaml:"naming,omitempty"`
		Empty     string
	}{}

	assert.Equal(t, "naming", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("JsonName"))))
	assert.Equal(t, "naming", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("JsonName2"))))
	assert.Equal(t, "naming", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("YamlName"))))
	assert.Equal(t, "naming", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("YamlName2"))))
	assert.Equal(t, "", TagNameFunction(mustFieldByName(reflect.TypeOf(s).FieldByName("Empty"))))
}

func mustFieldByName(st reflect.StructField, b bool) reflect.StructField {
	return st
}
