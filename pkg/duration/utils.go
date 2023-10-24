package duration

import "time"

func Min(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d2
	}
	return d1
}

func Max(d1 time.Duration, d2 time.Duration) time.Duration {
	if d2 > d1 {
		return d2
	}
	return d1
}
