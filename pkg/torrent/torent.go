package torrent

import (
	"context"
	"github.com/anacrolix/missinggo/v2"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

type Torrent struct {
	metaInfo metainfo.MetaInfo
	info     metainfo.Info
	closed   missinggo.SynchronizedEvent
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
