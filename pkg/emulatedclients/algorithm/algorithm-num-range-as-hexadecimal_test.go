package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/casing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestUnmarshalNumRangeAsHexadecimalAlgorithm(t *testing.T) {
	yamlString := `---
type: NUM_RANGE_ENCODED_AS_HEXADECIMAL
min: 1
max: 350
trimLeadingZeroes: true
maxLength: 8
case: lower
`
	algorithm := &Algorithm{}
	err := yaml.Unmarshal([]byte(yamlString), algorithm)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = algorithm.AfterPropertiesSet()
	assert.IsType(t, &NumRangeAsHexadecimalAlgorithm{}, algorithm.impl)
	assert.Equal(t, int64(1), algorithm.impl.(*NumRangeAsHexadecimalAlgorithm).Min)
	assert.Equal(t, int64(350), algorithm.impl.(*NumRangeAsHexadecimalAlgorithm).Max)
	assert.True(t, algorithm.impl.(*NumRangeAsHexadecimalAlgorithm).TrimLeadingZeroes)
	assert.Equal(t, 8, algorithm.impl.(*NumRangeAsHexadecimalAlgorithm).MaxLength)
	assert.Equal(t, casing.Lower, algorithm.impl.(*NumRangeAsHexadecimalAlgorithm).Case)
}

func TestDigitRangeAsHexadecimalAlgorithm_Generate(t *testing.T) {
	type fields struct {
		Min               int64
		Max               int64
		TrimLeadingZeroes bool
		maxLength         int
		Case              casing.Case
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "shouldFormatWithLeadingZeroes", fields: fields{Min: 1, Max: 1, TrimLeadingZeroes: false, maxLength: 8}, want: "00000001"},
		{name: "shouldFormatWithoutLeadingZeroes", fields: fields{Min: 1, Max: 1, TrimLeadingZeroes: true, maxLength: 8}, want: "1"},
		{name: "shouldFormatLargeNumbersAndTrimAccordingToLength", fields: fields{Min: 9223372036854775807, Max: 9223372036854775807, TrimLeadingZeroes: true, maxLength: 9}, want: "fffffffff"},
		{name: "shouldFormatLargeNumbersAndTrimAccordingToLength2", fields: fields{Min: 9223372036854775807, Max: 9223372036854775807, TrimLeadingZeroes: false, maxLength: 9}, want: "fffffffff"},
		{name: "shouldApplyCase", fields: fields{Min: 12, Max: 12, TrimLeadingZeroes: false, maxLength: 8, Case: casing.Lower}, want: "0000000c"},
		{name: "shouldApplyCase", fields: fields{Min: 12, Max: 12, TrimLeadingZeroes: false, maxLength: 8, Case: casing.Upper}, want: "0000000C"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &NumRangeAsHexadecimalAlgorithm{
				Min:               tt.fields.Min,
				Max:               tt.fields.Max,
				TrimLeadingZeroes: tt.fields.TrimLeadingZeroes,
				MaxLength:         tt.fields.maxLength,
				Case:              tt.fields.Case,
			}
			if got := a.Generate(); got != tt.want {
				t.Errorf("Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}
