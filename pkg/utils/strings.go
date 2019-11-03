package utils

import "time"

type AccessAwareString struct {
	lastAccessed time.Time
	val          string
}

func AccessAwareStringNew(str string) *AccessAwareString {
	return &AccessAwareString{
		lastAccessed: time.Now(),
		val:          str,
	}
}
func AccessAwareStringNewSince(str string, lastAccessed time.Time) *AccessAwareString {
	return &AccessAwareString{
		lastAccessed: lastAccessed,
		val:          str,
	}
}

func (s *AccessAwareString) Get() string {
	s.lastAccessed = time.Now()
	return s.val
}
func (s *AccessAwareString) LastAccess() time.Duration {
	return time.Now().Sub(s.lastAccessed)
}
