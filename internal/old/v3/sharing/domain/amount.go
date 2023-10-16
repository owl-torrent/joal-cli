package domain

import (
	"fmt"
)

type ByteAmount struct {
	amountInBytes int64
}

func NewByteAmount(bytes int64) (ByteAmount, error) {
	if bytes < 0 {
		return ByteAmount{}, fmt.Errorf("can not create a negative ByteAmount of [%d]", bytes)
	}
	return ByteAmount{amountInBytes: bytes}, nil
}

func (a *ByteAmount) add(amount ByteAmount) {
	a.amountInBytes += amount.amountInBytes
}

func (a *ByteAmount) GetInBytes() int64 {
	return a.amountInBytes
}
