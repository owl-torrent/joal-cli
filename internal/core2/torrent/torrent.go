package torrent

import (
	trackerlib "github.com/anacrolix/torrent/tracker"
)

type torrent struct {
	contrib  contribution
	peers    peersElector
	trackers trackerPool
}

func (t *torrent) GetPeers() Peers {
	return t.peers.GetPeers()
}

func (t *torrent) AnnounceStop(event trackerlib.AnnounceEvent, announcingFunction AnnouncingFunction) {
	//TODO implement me
	panic("implement me")
}

func (t *torrent) AnnounceToReadyTrackers(announcingFunction AnnouncingFunction) {
	//TODO implement me
	panic("implement me")
}

func (t *torrent) HandleAnnounceSuccess(response TrackerAnnounceResponse) {
	//TODO implement me
	panic("implement me")
}

func (t *torrent) HandleAnnounceError(response TrackerAnnounceResponseError) {
	//TODO implement me
	panic("implement me")
}
