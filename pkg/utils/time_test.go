package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEvery_shouldCallMethodAtIntervalUntilStop(t *testing.T) {
	duration := 1 * time.Millisecond
	counter := 0
	f := func() {
		counter++
	}

	quit := Every(duration, f)
	time.Sleep(100 * time.Millisecond)
	close(quit)
	assert.Greater(t, counter, 30) // check greater than 30 because running 100 times in 100 milliseconds is a bit too much
}

func TestEvery_shouldNotCallAfterQuitChannelWasClosed(t *testing.T) {
	duration := 1 * time.Millisecond
	counter := 0
	f := func() {
		counter++
	}

	quit := Every(duration, f)
	close(quit)
	counterAfterStop := counter
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, counter, counterAfterStop) // check greater than 30 because running 100 times in 100 milliseconds is a bit too much
}
