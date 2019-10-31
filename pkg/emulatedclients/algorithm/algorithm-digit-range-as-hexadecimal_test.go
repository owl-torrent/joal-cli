package algorithm

import (
	"testing"
)

func TestDigitRangeAsHexadecimalAlgorithm_Generate(t *testing.T) {
	type fields struct {
		Min               *int64
		Max               *int64
		TrimLeadingZeroes bool
		maxLength         *int
		Case              Case
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "shouldFormatWithLeadingZeroes", fields: fields{Min: i64p(1), Max: i64p(1), TrimLeadingZeroes: false, maxLength: i8p(8)}, want: "00000001"},
		{name: "shouldFormatWithoutLeadingZeroes", fields: fields{Min: i64p(1), Max: i64p(1), TrimLeadingZeroes: true, maxLength: i8p(8)}, want: "1"},
		{name: "shouldFormatLargeNumbersAndTrimAccordingToLength", fields: fields{Min: i64p(9223372036854775807), Max: i64p(9223372036854775807), TrimLeadingZeroes: true, maxLength: i8p(9)}, want: "fffffffff"},
		{name: "shouldFormatLargeNumbersAndTrimAccordingToLength2", fields: fields{Min: i64p(9223372036854775807), Max: i64p(9223372036854775807), TrimLeadingZeroes: false, maxLength: i8p(9)}, want: "fffffffff"},
		{name: "shouldApplyCase", fields: fields{Min: i64p(12), Max: i64p(12), TrimLeadingZeroes: false, maxLength: i8p(8), Case: Lower}, want: "0000000c"},
		{name: "shouldApplyCase", fields: fields{Min: i64p(12), Max: i64p(12), TrimLeadingZeroes: false, maxLength: i8p(8), Case: Upper}, want: "0000000C"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &NumRangeAsHexadecimalAlgorithm{
				Min:               tt.fields.Min,
				Max:               tt.fields.Max,
				TrimLeadingZeroes: tt.fields.TrimLeadingZeroes,
				MaxLength:         tt.fields.maxLength,
				Case:              &tt.fields.Case,
			}
			if got := a.Generate(); got != tt.want {
				t.Errorf("Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func i64p(v int64) *int64 {
	return &v
}
func i8p(v int) *int {
	return &v
}