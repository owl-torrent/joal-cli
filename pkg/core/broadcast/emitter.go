package broadcast

func EmitSeedStart(event SeedStartedEvent) {
	listeners.onSeedStart(event)
}

func EmitSeedStop(event SeedStoppedEvent) {
	listeners.onSeedStop(event)
}

func EmitConfigChanged(event ConfigChangedEvent) {
	listeners.onConfigChanged(event)
}

func EmitTorrentAdded(event TorrentAddedEvent) {
	listeners.onTorrentAdded(event)
}

func EmitTorrentAnnouncing(event TorrentAnnouncingEvent) {
	listeners.onTorrentAnnouncing(event)
}

func EmitTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent) {
	listeners.onTorrentAnnounceSuccess(event)
}

func EmitTorrentAnnounceFailed(event TorrentAnnounceFailedEvent) {
	listeners.onTorrentAnnounceFailed(event)
}

func EmitTorrentSwarmChanged(event TorrentSwarmChangedEvent) {
	listeners.onTorrentSwarmChanged(event)
}

func EmitTorrentRemoved(event TorrentRemovedEvent) {
	listeners.onTorrentRemoved(event)
}

func EmitNoticeableError(event NoticeableErrorEvent) {
	listeners.onNoticeableError(event)
}

func EmitGlobalBandwidthChanged(event GlobalBandwidthChangedEvent) {
	listeners.onGlobalBandwidthChanged(event)
}

func EmitBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent) {
	listeners.onBandwidthWeightHasChanged(event)
}
