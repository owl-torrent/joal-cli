package torrent

type torrent struct {
	contrib  contribution
	peers    peersElector
	trackers trackerPool
}

func (t *torrent) GetPeers() Peers {
	return t.peers.GetPeers()
}

func (t *torrent) AnnounceStop(announcingFunction AnnouncingFunction) {
	//TODO: implement me
	panic("implement me")
}

func (t *torrent) AnnounceToReadyTrackers(announcingFunction AnnouncingFunction) {
	//TODO: implement me
	panic("implement me")
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
