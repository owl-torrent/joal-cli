package torrent

type contribution struct {
	uploaded   int64
	downloaded int64
	left       int64
	corrupt    int64
}

func (c *contribution) addUploaded(bytes int64) {
	if bytes < 0 {
		return
	}
	c.uploaded += bytes
}
