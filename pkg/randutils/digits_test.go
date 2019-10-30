package randutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRange(t *testing.T) {
	type args struct {
		minInclusive int64
		maxInclusive int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{name: "shouldWorkOnRange1", args: args{minInclusive: 1, maxInclusive: 1}, want: 1},
		{name: "shouldWorkOnRange1WithValue0", args: args{minInclusive: 0, maxInclusive: 0}, want: 0},
		{name: "shouldWorkOnRange1WithValue-1", args: args{minInclusive: -1, maxInclusive: -1}, want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Range(tt.args.minInclusive, tt.args.maxInclusive); got != tt.want {
				t.Errorf("Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeNegativeMinPositiveMax(t *testing.T) {
	for i := 0; i < 500; i++ {
		min := int64(-50)
		max := int64(i)
		actual := Range(min, max)
		assert.LessOrEqual(t, actual, max)
		assert.GreaterOrEqual(t, actual, min)
	}
}

func TestRangeNegativeMinNegativeMax(t *testing.T) {
	for i := 0; i < 500; i++ {
		min := int64(-50000)
		max := int64(-i)
		actual := Range(min, max)
		assert.LessOrEqual(t, actual, max)
		assert.GreaterOrEqual(t, actual, min)
	}
}
