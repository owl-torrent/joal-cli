package generator

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAccessAwareString_GetShouldRefreshLastAccess(t *testing.T) {
	aas := AccessAwarePeerIdNew([20]byte{0x25})
	aas.lastAccessed = time.Now().Add(-1 * time.Hour) // offset last access

	assert.Equal(t, peerid.PeerId([20]byte{0x25}), aas.Get())
	assert.Less(t, aas.LastAccess().Minutes(), float64(1)) // last access was refreshed and is less than 1m (initial value was 60 min)
}

func TestAccessAwareString_AccessAwareStringNewSince(t *testing.T) {
	expectedTime := time.Now().Add(-80 * time.Minute)
	aas := AccessAwarePeerIdNewSince([20]byte{0xff}, expectedTime)

	assert.Greater(t, aas.LastAccess().Milliseconds(), (79 * time.Minute).Milliseconds()) // last access was refreshed and is less than 1m (initial value was 60 min)
	assert.Equal(t, peerid.PeerId([20]byte{0xff}), aas.Get())
	assert.Less(t, aas.LastAccess().Minutes(), float64(1))
}
