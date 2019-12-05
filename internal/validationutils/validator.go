package validationutils

import (
	"reflect"
	"strings"
)

var TagNameFunction = func(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
	if name == "" {
		name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	}
	if name == "-" {
		return ""
	}
	return name
}
