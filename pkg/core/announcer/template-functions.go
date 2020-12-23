package announcer

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/casing"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/urlencoder"
	"strconv"
	"text/template"
)

func TemplateFunctions(encoder *urlencoder.UrlEncoder) template.FuncMap {
	return template.FuncMap{
		"byteArray20ToString": func(s [20]byte) string {
			return string(s[:])
		},
		"uint32ToHexString": func(k uint32) string {
			hex := strconv.FormatInt(int64(k), 16)
			return hex
		},
		"withLeadingZeroes": func(str string, upToLength int) string {
			return fmt.Sprintf("%0"+strconv.Itoa(upToLength)+"s", str)
		},
		"toLower": func(str string) string {
			return casing.Lower.ApplyCase(str)
		},
		"toUpper": func(str string) string {
			return casing.Upper.ApplyCase(str)
		},
		"urlEncode": func(str string) string {
			return encoder.Encode(str)
		},
	}
}
