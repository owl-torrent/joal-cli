package broadcast

func EmitSeedStart(event SeedStartedEvent) {
	listeners.OnSeedStart(event)
}

func EmitSeedStop(event SeedStoppedEvent) {
	listeners.OnSeedStop(event)
}

func EmitConfigChanged(event ConfigChangedEvent) {
	listeners.OnConfigChanged(event)
}

func EmitTorrentAdded(event TorrentAddedEvent) {
	listeners.OnTorrentAdded(event)
}

func EmitTorrentAnnouncing(event TorrentAnnouncingEvent) {
	listeners.OnTorrentAnnouncing(event)
}

func EmitTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent) {
	listeners.OnTorrentAnnounceSuccess(event)
}

func EmitTorrentAnnounceFailed(event TorrentAnnounceFailedEvent) {
	listeners.OnTorrentAnnounceFailed(event)
}

func EmitTorrentSwarmChanged(event TorrentSwarmChangedEvent) {
	listeners.OnTorrentSwarmChanged(event)
}

func EmitTorrentRemoved(event TorrentRemovedEvent) {
	listeners.OnTorrentRemoved(event)
}

func EmitNoticeableError(event NoticeableErrorEvent) {
	listeners.OnNoticeableError(event)
}

func EmitGlobalBandwidthChanged(event GlobalBandwidthChangedEvent) {
	listeners.OnGlobalBandwidthChanged(event)
}

func EmitBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent) {
	listeners.OnBandwidthWeightHasChanged(event)
}
