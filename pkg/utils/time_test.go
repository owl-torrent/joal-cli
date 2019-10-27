package utils

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestEvery_shouldCallMethodAtIntervalUntilStop(t *testing.T) {
	wg := sync.WaitGroup{}
	duration := 1 * time.Millisecond
	counter := 0
	f := func() {
		counter++
		wg.Done()
	}

	wg.Add(3)
	quit := Every(duration, f)
	wg.Wait()
	close(quit)
	assert.GreaterOrEqual(t, counter, 3) // check greater than 30 because running 100 times in 100 milliseconds is a bit too much
}
