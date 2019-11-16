package key

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKey_Format(t *testing.T) {
	formatter := func (k Key) string { return "coucou" }
	assert.Equal(t, "coucou", Key(546).Format(formatter))
}