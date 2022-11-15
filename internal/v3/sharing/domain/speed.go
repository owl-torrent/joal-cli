package domain

import (
	"fmt"
	"time"
)

type Speed struct {
	bps int64
}

func (s *Speed) recalculateSpeed(bytes ByteAmount, timespan time.Duration) error {
	if timespan.Milliseconds() < 0 {
		return fmt.Errorf("can not calculate speed for negative duration, [%s] received", timespan)
	}

	if bytes.GetInBytes() == 0 {
		s.bps = 0
		return nil
	}

	s.bps = bytes.GetInBytes() * int64(timespan.Seconds())
	return nil
}

func (s *Speed) BytesPerSeconds() int64 {
	return s.bps
}
