package domain

import "time"

type DownloadStat struct {
	Speed         Speed
	Amount        ByteAmount
	lastTimeAdded time.Time
}

func (d *DownloadStat) add(amount ByteAmount) error {
	deltaBetweenLastAdd := time.Now().Sub(d.lastTimeAdded)

	d.lastTimeAdded = time.Now()
	d.Amount.add(amount)
	return d.Speed.recalculateSpeed(amount, deltaBetweenLastAdd)
}
