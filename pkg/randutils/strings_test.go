package randutils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	type args struct {
		runes  string
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "shouldGenerateWithSmallSet", args: args{runes: "a", length: 10}, want: "aaaaaaaaaa"},
		{name: "shouldGenerateWithSmallSetAndLargeLength", args: args{runes: "a", length: 1000}, want: strings.Repeat("a", 1000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.args.runes, tt.args.length); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringUniformDistribution(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	actualDistribution := make(map[uint8]int64, len(charset))
	for i := 0; i < len(charset); i++ {
		actualDistribution[charset[i]] = 0
	}

	const (
		iterations       = 500
		randStringLength = 500
	)

	for i := 0; i < iterations; i++ {
		str := String(charset, randStringLength)
		for j := 0; j < len(str); j++ {
			actualDistribution[str[j]] = actualDistribution[str[j]] + 1
		}
	}

	expectedDistribution := make(map[uint8]int64, len(charset))
	for i := 0; i < len(charset); i++ {
		expectedDistribution[charset[i]] = (iterations * randStringLength) / int64(len(charset))
	}

	acceptedDeviation := float64((iterations*randStringLength)/int64(len(charset))) * 0.1 // 8%

	assert.InDeltaMapValues(t, expectedDistribution, actualDistribution, acceptedDeviation)
}
