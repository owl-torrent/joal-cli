package torrent

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContribution_AddUploaded(t *testing.T) {
	var c contribution

	c.addUploaded(25)
	assert.Equal(t, int64(25), c.uploaded)

	c.addUploaded(11111)
	assert.Equal(t, int64(11136), c.uploaded)
}

func TestContribution_AddUploadedShouldPreventAddingNegativeNumber(t *testing.T) {
	var c contribution

	c.addUploaded(-20)

	assert.Equal(t, int64(0), c.uploaded)
}
