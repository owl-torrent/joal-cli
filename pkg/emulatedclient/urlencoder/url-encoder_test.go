package urlencoder

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient/casing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestUrlEncoder_ShouldUnmarshalWithoutExcludedChars(t *testing.T) {
	yamlString := `---
encodedHexCase: lower
`
	urlEncoder := &UrlEncoder{}
	err := yaml.Unmarshal([]byte(yamlString), urlEncoder)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	assert.Equal(t, casing.Lower, urlEncoder.EncodedHexCase)
}

func TestUrlEncoder_Encode(t *testing.T) {
	type fields struct {
		EncodedHexCase casing.Case
	}
	type args struct {
		str string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "ShouldUrlEncode", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: "i love potatoes"}, want: "i%20love%20potatoes"},
		{name: "ShouldNotConvertCaseOfNonEncodedChars", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: "iLovePotatoes"}, want: "iLovePotatoes"},
		{name: "ShouldNotConvertCaseOfNonEncodedChars", fields: fields{EncodedHexCase: casing.Upper}, args: args{str: "iLovePotatoes"}, want: "iLovePotatoes"},
		{name: "ShouldConvertCaseOfNonEncodedChars", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: "4Ngry-pümk|n"}, want: "4Ngry-p%c3%bcmk%7cn"},
		{name: "ShouldConvertCaseOfNonEncodedChars", fields: fields{EncodedHexCase: casing.Upper}, args: args{str: "4Ngry-pümk|n"}, want: "4Ngry-p%C3%BCmk%7Cn"},
		{name: "ShouldConvertCaseOfNonEncodedChars", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: "+1"}, want: "%2b1"},
		{name: "ShouldUrlEncode", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: string(rune(0x00))}, want: "%00"},
		{name: "ShouldUrlEncode", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: string(rune(0x01))}, want: "%01"},
		{name: "ShouldUrlEncode", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: string(rune(0x10))}, want: "%10"},
		{name: "ShouldUrlEncode", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: string(rune(0x1e))}, want: "%1e"},
		{name: "ShouldUrlEncode", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: string(rune(0x20))}, want: "%20"},
		{name: "ShouldUrlEncode", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: string(rune(0x1A))}, want: "%1a"},
		{name: "ShouldRespectRFC3986", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~!#$%&'()*+,/:;=?@[]"}, want: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~%21%23%24%25%26%27%28%29%2a%2b%2c%2f%3a%3b%3d%3f%40%5b%5d"},
		{name: "ShouldRespectRFC3986", fields: fields{EncodedHexCase: casing.Lower}, args: args{str: createStrFrom00toFF()}, want: "%00%01%02%03%04%05%06%07%08%09%0a%0b%0c%0d%0e%0f%10%11%12%13%14%15%16%17%18%19%1a%1b%1c%1d%1e%1f%20%21%22%23%24%25%26%27%28%29%2a%2b%2c-.%2f0123456789%3a%3b%3c%3d%3e%3f%40ABCDEFGHIJKLMNOPQRSTUVWXYZ%5b%5c%5d%5e_%60abcdefghijklmnopqrstuvwxyz%7b%7c%7d~%7f%80%81%82%83%84%85%86%87%88%89%8a%8b%8c%8d%8e%8f%90%91%92%93%94%95%96%97%98%99%9a%9b%9c%9d%9e%9f%a0%a1%a2%a3%a4%a5%a6%a7%a8%a9%aa%ab%ac%ad%ae%af%b0%b1%b2%b3%b4%b5%b6%b7%b8%b9%ba%bb%bc%bd%be%bf%c0%c1%c2%c3%c4%c5%c6%c7%c8%c9%ca%cb%cc%cd%ce%cf%d0%d1%d2%d3%d4%d5%d6%d7%d8%d9%da%db%dc%dd%de%df%e0%e1%e2%e3%e4%e5%e6%e7%e8%e9%ea%eb%ec%ed%ee%ef%f0%f1%f2%f3%f4%f5%f6%f7%f8%f9%fa%fb%fc%fd%fe%ff"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UrlEncoder{
				EncodedHexCase: tt.fields.EncodedHexCase,
			}
			if got := u.Encode(tt.args.str); got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createStrFrom00toFF() string {
	sb := strings.Builder{}
	sb.Grow(255)
	for i := 0x00; i <= 0xff; i++ {
		sb.WriteByte(byte(i))
	}
	return sb.String()
}
