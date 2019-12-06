package seed

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_SeedStats_AddUploaded(t *testing.T) {
	stats := seedStats{
		Downloaded: 0,
		Left:       0,
		Uploaded:   0,
	}

	stats.AddUploaded(50)
	assert.Equal(t, int64(0), stats.Downloaded)
	assert.Equal(t, int64(0), stats.Left)
	assert.Equal(t, int64(50), stats.Uploaded)

	stats.AddUploaded(250)
	assert.Equal(t, int64(0), stats.Downloaded)
	assert.Equal(t, int64(0), stats.Left)
	assert.Equal(t, int64(300), stats.Uploaded)
}
func Test_SeedStats_ResetUploaded(t *testing.T) {
	stats := seedStats{
		Downloaded: 200,
		Left:       250,
		Uploaded:   960,
	}

	stats.ResetUploaded()
	assert.Equal(t, int64(200), stats.Downloaded)
	assert.Equal(t, int64(250), stats.Left)
	assert.Equal(t, int64(0), stats.Uploaded)
}
