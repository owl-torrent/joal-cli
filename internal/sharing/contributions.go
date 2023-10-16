package sharing

type Contribution struct {
	uploaded   int64
	downloaded int64
	left       int64
}

func (c Contribution) addUpload(amount int64) Contribution {
	c.uploaded += amount
	return c
}

func (c Contribution) addDownload(amount int64) Contribution {
	if amount > c.left {
		c.downloaded += c.left
		c.left = 0
		return c
	}
	c.downloaded += amount
	c.left -= amount
	return c
}

func (c Contribution) isDownloadComplete() bool {
	return c.left == 0
}

type Contributions struct {
	overall Contribution
	session Contribution
}

func (c Contributions) addUpload(amount int64) Contributions {
	c.session = c.session.addUpload(amount)
	c.overall = c.overall.addUpload(amount)
	return c
}

func (c Contributions) addDownload(amount int64) Contributions {
	c.session = c.session.addDownload(amount)
	c.overall = c.overall.addDownload(amount)
	return c
}

func (c Contributions) isDownloadComplete() bool {
	return c.session.isDownloadComplete()
}
