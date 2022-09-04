package torrent

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContribution_AddUploaded(t *testing.T) {
	var c Contribution

	c.AddUploaded(25)
	assert.Equal(t, int64(25), c.uploaded)

	c.AddUploaded(11111)
	assert.Equal(t, int64(11136), c.uploaded)
}

func TestContribution_AddUploadedShouldPreventAddingNegativeNumber(t *testing.T) {
	var c Contribution

	c.AddUploaded(-20)

	assert.Equal(t, int64(0), c.uploaded)
}
