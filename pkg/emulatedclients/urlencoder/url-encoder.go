package urlencoder

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/casing"
	"net/url"
	"strings"
)

type UrlEncoder struct {
	EncodedHexCase casing.Case `yaml:"encodedHexCase"`
}

func (u *UrlEncoder) Encode(str string) string {
	encoded := url.QueryEscape(str)
	encoded = strings.ReplaceAll(encoded, "+", "%20")

	sb := &strings.Builder{}
	sb.Grow(len(encoded))
	for i := 0; i < len(encoded); i++ {
		currentChar := encoded[i]
		sb.WriteByte(currentChar)
		if encoded[i] == '%' {
			sb.WriteString(u.EncodedHexCase.ApplyCase(encoded[i+1 : i+3]))
			i += 2
		}
	}
	return sb.String()
}
