package gomockutils

import "fmt"

type greaterThanMatcher struct {
	x int64
}

func (m *greaterThanMatcher) Matches(x interface{}) bool {
	val, ok := x.(int64)
	if !ok {
		return false
	}
	return val > m.x
}

// String describes what the matcher matches.
func (m *greaterThanMatcher) String() string {
	return fmt.Sprintf("is greater than %d", m.x)
}

func NewGreaterThanMatcher(x int64) *greaterThanMatcher {
	return &greaterThanMatcher{x: x}
}
