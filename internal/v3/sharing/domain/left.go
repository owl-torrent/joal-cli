package domain

type LeftStat struct {
	bytes int64
}

func (l *LeftStat) subtract(amount ByteAmount) {
	l.bytes -= amount.GetInBytes()
	if l.bytes < 0 {
		l.bytes = 0
	}
}

func (l *LeftStat) Amount() ByteAmount {
	return ByteAmount{amountInBytes: l.bytes}
}
