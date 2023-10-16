package bandwidth

import (
	"github.com/anthonyraymond/joal-cli/internal/old/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomSpeedProvider_ShouldBuildFromConfig(t *testing.T) {
	rsp := newRandomSpeedProvider(&core.SpeedProviderConfig{
		MinimumBytesPerSeconds: 10,
		MaximumBytesPerSeconds: 100,
	})

	assert.Equal(t, int64(10), rsp.(*randomSpeedProvider).MinimumBytesPerSeconds)
	assert.Equal(t, int64(100), rsp.(*randomSpeedProvider).MaximumBytesPerSeconds)
}

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
			r := &randomSpeedProvider{
				MinimumBytesPerSeconds: tt.fields.minimumBytesPerSeconds,
				MaximumBytesPerSeconds: tt.fields.maximumBytesPerSeconds,
				value:                  tt.fields.value,
			}
			for i := 0; i < 100; i++ {
				r.Refresh()
				bps := r.GetBytesPerSeconds()
				assert.GreaterOrEqual(t, bps, r.MinimumBytesPerSeconds)
				assert.LessOrEqual(t, bps, r.MaximumBytesPerSeconds)
			}
		})
	}
}

func TestRandomSpeedProviderRefreshShouldGenerateRandomValues(t *testing.T) {
	r := randomSpeedProvider{
		MinimumBytesPerSeconds: 20,
		MaximumBytesPerSeconds: 100000,
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
