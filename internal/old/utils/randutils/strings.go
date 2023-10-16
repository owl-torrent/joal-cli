package randutils

import "strings"

func String(chars string, length int) string {
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(chars[globalRand.Intn(len(chars))])
	}
	return sb.String()
}
