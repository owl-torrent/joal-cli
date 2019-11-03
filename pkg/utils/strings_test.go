package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAccessAwareString_GetShouldRefreshLastAccess(t *testing.T) {
	aas := AccessAwareStringNew("test")
	aas.lastAccessed = time.Now().Add(-1 * time.Hour) // offset last access

	assert.Equal(t, "test", aas.Get())
	assert.Less(t, aas.LastAccess().Minutes(), float64(1)) // last access was refreshed and is less than 1m (initial value was 60 min)
}

func TestAccessAwareString_AccessAwareStringNewSince(t *testing.T) {
	expectedTime := time.Now().Add(-80 * time.Minute)
	aas := AccessAwareStringNewSince("test", expectedTime)

	assert.Greater(t, aas.LastAccess().Milliseconds(), (79 * time.Minute).Milliseconds()) // last access was refreshed and is less than 1m (initial value was 60 min)
	assert.Equal(t, "test", aas.Get())
	assert.Less(t, aas.LastAccess().Minutes(), float64(1))
}
