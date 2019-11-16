package peerid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPeerId_Format(t *testing.T) {
	formatter := func (p PeerId) string { return "coucou" }
	assert.Equal(t, "coucou", PeerId([20]byte{0x01}).Format(formatter))
}