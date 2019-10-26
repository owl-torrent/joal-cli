package bandwidth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomSpeedProvider_RefreshShouldGenerateValueWithinRange(t *testing.T) {
	type fields struct {
		minimumBytesPerSeconds int64
		maximumBytesPerSeconds int64
		value                  int64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "withLargeRange", fields: fields{minimumBytesPerSeconds: 2000, maximumBytesPerSeconds: 1000000}},
		{name: "withLargeRange2", fields: fields{minimumBytesPerSeconds: 20000000, maximumBytesPerSeconds: 1000000000}},
		{name: "withSmallRange", fields: fields{minimumBytesPerSeconds: 0, maximumBytesPerSeconds: 1}},
		{name: "withSmallRange2", fields: fields{minimumBytesPerSeconds: 50, maximumBytesPerSeconds: 51}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RandomSpeedProvider{
				minimumBytesPerSeconds: tt.fields.minimumBytesPerSeconds,
				maximumBytesPerSeconds: tt.fields.maximumBytesPerSeconds,
				value:                  tt.fields.value,
			}
			for i := 0; i < 1000; i++ {
				r.Refresh()
				bps := r.GetBytesPerSeconds()
				assert.GreaterOrEqual(t, bps, r.minimumBytesPerSeconds)
				assert.LessOrEqual(t, bps, r.maximumBytesPerSeconds)
			}
		})
	}
}

func TestRandomSpeedProviderRefreshShouldGenerateRandomValues(t *testing.T) {
	r := RandomSpeedProvider{
		minimumBytesPerSeconds: 20,
		maximumBytesPerSeconds: 100000,
	}

	// One key for each unique value
	valueSet := make(map[int64]bool)

	for i := 0; i < 1000; i++ {
		r.Refresh()
		valueSet[r.GetBytesPerSeconds()] = true
	}

	// Must ave more than 1 key if generated values were different
	assert.Greater(t, len(valueSet), 1)
}
