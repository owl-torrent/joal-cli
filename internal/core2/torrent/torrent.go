package torrent

import (
	libtorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	libtracker "github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"go.uber.org/zap"
	"time"
)

type torrent struct {
	infoHash libtorrent.InfoHash
	info     slimInfo
	contrib  contribution
	peers    peersElector
	trackers trackerPool
}

func (t *torrent) Name() string {
	return t.info.name
}

func (t *torrent) InfoHash() libtorrent.InfoHash {
	return t.infoHash
}

func (t *torrent) GetPeers() Peers {
	return t.peers.GetPeers()
}

func (t *torrent) AnnounceStop(announcingFunction AnnouncingFunction) {
	logger := logs.GetLogger().With(zap.String("torrent", t.info.name))

	trackersInUse := t.trackers.inUse()
	for _, tr := range trackersInUse {
		err := tr.announce(libtracker.Stopped, withRequestEnhancer(announcingFunction, *t))
		if err != nil {
			logger.Error("failed to announce for tracker ", zap.String("tracker", tr.url.String()), zap.Error(err))
		}
	}
}

func (t *torrent) AnnounceToReadyTrackers(announcingFunction AnnouncingFunction) {
	logger := logs.GetLogger().With(zap.String("torrent", t.info.name))

	trackerReady := t.trackers.readyToAnnounce(time.Now())
	for _, tr := range trackerReady {
		err := tr.announce(libtracker.None, withRequestEnhancer(announcingFunction, *t))
		if err != nil {
			logger.Error("failed to announce for tracker ", zap.String("tracker", tr.url.String()), zap.Error(err))
		}
	}
}

func withRequestEnhancer(baseFunc AnnouncingFunction, t torrent) AnnouncingFunction {
	return func(request TrackerAnnounceRequest) {
		request.InfoHash = t.infoHash
		request.Uploaded = t.contrib.uploaded
		request.Downloaded = t.contrib.downloaded
		request.Left = t.contrib.left
		request.Corrupt = t.contrib.corrupt
		request.Private = t.info.private

		baseFunc(request)
	}
}

func (t *torrent) HandleAnnounceSuccess(response TrackerAnnounceResponse) {
	trackerUrl := response.Request.Url
	t.trackers.succeed(trackerUrl, response)

	t.peers.updatePeersForTracker(peersUpdateRequest{
		trackerUrl: trackerUrl,
		Seeders:    response.Seeders,
		Leechers:   response.Leechers,
	})
}

func (t *torrent) HandleAnnounceError(response TrackerAnnounceResponseError) {
	trackerUrl := response.Request.Url
	t.trackers.failed(trackerUrl, response)

	t.peers.removePeersForTracker(peersDeleteRequest{trackerUrl: trackerUrl})
}

type slimInfo struct {
	pieceLength int64
	name        string
	length      int64
	private     bool
	source      string
}

func infoToSlimInfo(info metainfo.Info) slimInfo {
	si := slimInfo{
		pieceLength: info.PieceLength,
		name:        info.Name,
		length:      info.Length,
		private:     false,
		source:      info.Source,
	}
	if info.Private != nil {
		si.private = *info.Private
	}
	return si
}
