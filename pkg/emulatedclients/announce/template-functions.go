package announce

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/casing"
	"strconv"
	"text/template"
)

var templateFunctions template.FuncMap

func init() {
	templateFunctions = template.FuncMap{
		"uint32ToHexString": func(k uint32) string {
			hex := strconv.FormatInt(int64(k), 16)
			return fmt.Sprintf("%s", hex)
		},
		"withLeadingZeroes": func(str string, upToLength int) string {
			return fmt.Sprintf("%0" + strconv.Itoa(upToLength) + "s", str)
		},
		"toLower": func(str string) string {
			return casing.Lower.ApplyCase(str)
		},
		"toUpper": func(str string) string {
			return casing.Upper.ApplyCase(str)
		},
	}
}