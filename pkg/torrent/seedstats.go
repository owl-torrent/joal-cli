package torrent

type seedStats struct {
	Downloaded int64
	Left       int64 // If less than 0, math.MaxInt64 will be used for HTTP trackers instead.
	Uploaded   int64
}

func (ss *seedStats) AddUploaded(bytes int64) {
	ss.Uploaded += bytes
}
func (ss *seedStats) ResetUploaded() {
	ss.Uploaded = 0
}