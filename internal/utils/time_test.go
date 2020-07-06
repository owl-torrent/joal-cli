package utils

import (
	"github.com/nvn1729/congo"
	"testing"
	"time"
)

func TestEvery_shouldCallMethodAtIntervalUntilStop(t *testing.T) {
	duration := 1 * time.Millisecond
	latch := congo.NewCountDownLatch(3)
	f := func() {
		_ = latch.CountDown()
	}

	quit := Every(duration, f)
	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch has timed out")
	}
	close(quit)
}
