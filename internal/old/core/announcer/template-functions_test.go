package announcer

import (
	"bytes"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/casing"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/urlencoder"
	"math"
	"testing"
	"text/template"
)

func Test_TemplateFunctions(t *testing.T) {
	urlEncoder := urlencoder.UrlEncoder{EncodedHexCase: casing.Upper}
	type args struct {
		templateString string
		templateData   map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "uint32ToHexString", args: args{templateString: "jo={{uint32ToHexString .Attr}}", templateData: map[string]interface{}{"Attr": uint32(254)}}, want: "jo=fe"},
		{name: "uint32ToHexString", args: args{templateString: "jo={{uint32ToHexString .Attr}}", templateData: map[string]interface{}{"Attr": uint32(0)}}, want: "jo=0"},
		{name: "uint32ToHexString", args: args{templateString: "jo={{uint32ToHexString .Attr}}", templateData: map[string]interface{}{"Attr": uint32(math.MaxUint32)}}, want: "jo=ffffffff"},
		{name: "byteArrayToString", args: args{templateString: "jo={{byteArray20ToString .Attr}}", templateData: map[string]interface{}{"Attr": [20]byte{0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61}}}, want: "jo=aaaaaaaaaaaaaaaaaaaa"},
		{name: "byteArrayToString", args: args{templateString: "jo={{byteArray20ToString .Attr}}", templateData: map[string]interface{}{"Attr": [20]byte{0x61, 0x00, 0x61, 0xff, 0x10, 0x61, 0x00, 0x61, 0xff, 0x10, 0x61, 0x00, 0x61, 0xff, 0x10, 0x61, 0x00, 0x61, 0xff, 0x10}}}, want: "jo=a\x00a\xff\x10a\x00a\xff\x10a\x00a\xff\x10a\x00a\xff\x10"},
		{name: "byteArrayToString", args: args{templateString: "jo={{byteArray20ToString .Attr}}", templateData: map[string]interface{}{"Attr": [20]byte{}}}, want: "jo=\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"},
		{name: "withLeadingZeroes", args: args{templateString: "jo={{withLeadingZeroes .Attr 8}}", templateData: map[string]interface{}{"Attr": ""}}, want: "jo=00000000"},
		{name: "withLeadingZeroes", args: args{templateString: "jo={{withLeadingZeroes .Attr 8}}", templateData: map[string]interface{}{"Attr": "123"}}, want: "jo=00000123"},
		{name: "withLeadingZeroes", args: args{templateString: "jo={{withLeadingZeroes .Attr 8}}", templateData: map[string]interface{}{"Attr": "123456789"}}, want: "jo=123456789"},
		{name: "toLower", args: args{templateString: "jo={{toLower .Attr}}", templateData: map[string]interface{}{"Attr": "AbcD"}}, want: "jo=abcd"},
		{name: "toLower", args: args{templateString: "jo={{toLower .Attr}}", templateData: map[string]interface{}{"Attr": "abcd"}}, want: "jo=abcd"},
		{name: "toUpper", args: args{templateString: "jo={{toUpper .Attr}}", templateData: map[string]interface{}{"Attr": "AbcD"}}, want: "jo=ABCD"},
		{name: "toUpper", args: args{templateString: "jo={{toUpper .Attr}}", templateData: map[string]interface{}{"Attr": "ABCD"}}, want: "jo=ABCD"},
		{name: "urlEncode", args: args{templateString: "jo={{urlEncode .Attr}}", templateData: map[string]interface{}{"Attr": "AB&%CD"}}, want: "jo=AB%26%25CD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryStringTemplate, err := template.New("mulit-func").Funcs(TemplateFunctions(&urlEncoder)).Parse(tt.args.templateString)
			if err != nil {
				panic(err)
			}
			writer := bytes.NewBufferString("")
			err = queryStringTemplate.Execute(writer, tt.args.templateData)
			if err != nil {
				t.Error(err)
			}

			got := writer.String()
			if got != tt.want {
				t.Errorf("setupQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
