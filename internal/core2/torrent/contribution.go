package torrent

type Contribution struct {
	uploaded   int64
	downloaded int64
	left       int64
	corrupt    int64
}

func (c *Contribution) AddUploaded(bytes int64) {
	if bytes < 0 {
		return
	}
	c.uploaded += bytes
}
