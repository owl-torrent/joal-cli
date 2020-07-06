package seed

import (
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

type ITorrent interface {
	InfoHash() torrent.InfoHash
	AddUploaded(bytes int64)
	// May return nil
	GetSwarm() ISwarm
	StartSeeding()
	StopSeeding()
}

type Torrent struct {
	metaInfo metainfo.MetaInfo
	info     metainfo.Info
}

type trackerAnnounceResult struct {
	Err       error
	Interval  time.Duration
	Completed time.Time
}

func (t *Torrent) announce(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (ret trackerAnnounceResult) {
	defer func() {
		ret.Completed = time.Now()
	}()

	ret.Interval = 5 * time.Minute

	//*
	var res tracker.AnnounceResponse
	/*/
	TODO: client announce (with a lock??)
	if err != nil {
		ret.Err = fmt.Errorf("error announcing: %s", err)
		return
	}
	//*/

	ret.Interval = time.Duration(res.Interval) * time.Second
	return
}
