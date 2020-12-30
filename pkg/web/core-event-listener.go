package web

import (
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/pkg/core/broadcast"
	"sync"
)

const announceHistoryMaxLength = 5

type AppStateCoreListener struct {
	state *State
	lock  *sync.Mutex
}

func (l *AppStateCoreListener) onSeedStart(event broadcast.SeedStartedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.state.Started = true
	l.state.Client = &Client{
		Name:    event.Client,
		Version: event.Version,
	}

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onSeedStop(_ broadcast.SeedStoppedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.state = &State{
		Started:   false,
		Client:    nil,
		Config:    nil,
		Torrents:  nil,
		Bandwidth: nil,
	}

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onConfigChanged(event broadcast.ConfigChangedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.state.Config = &Config{
		NeedRestartToTakeEffect: event.NeedRestartToTakeEffect,
		RuntimeConfig: &RuntimeConfig{
			MinimumBytesPerSeconds: event.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds,
			MaximumBytesPerSeconds: event.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds,
			Client:                 event.RuntimeConfig.Client,
		},
	}

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onTorrentAdded(event broadcast.TorrentAddedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.state.Torrents == nil {
		l.state.Torrents = map[string]*Torrent{}
	}
	t := &Torrent{
		Infohash: event.Infohash.String(),
		Name:     event.Name,
		File:     event.File,
		Size:     event.Size,
		Uploaded: 0,
		Trackers: map[string]*TorrentTrackers{},
	}
	for _, u := range event.TrackerAnnounceUrls {
		t.Trackers[u.String()] = &TorrentTrackers{
			Url:             u,
			IsAnnouncing:    false,
			InUse:           false,
			AnnounceHistory: []*AnnounceResult{},
		}
	}

	l.state.Torrents[event.Infohash.String()] = t

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onTorrentAnnouncing(event broadcast.TorrentAnnouncingEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	l.state.Torrents[event.Infohash.String()].Uploaded = event.Uploaded

	tr := l.state.Torrents[event.Infohash.String()].Trackers[event.TrackerUrl]
	tr.InUse = true
	tr.IsAnnouncing = true

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onTorrentAnnounceSuccess(event broadcast.TorrentAnnounceSuccessEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	tr := l.state.Torrents[event.Infohash.String()].Trackers[event.TrackerUrl]
	tr.IsAnnouncing = false
	tr.Interval = int(event.Interval.Seconds())
	tr.Seeders = event.Seeder
	tr.Leechers = event.Leechers

	history := make([]*AnnounceResult, announceHistoryMaxLength)
	history[0] = &AnnounceResult{
		AnnounceEvent: event.AnnounceEvent,
		WasSuccessful: true,
		Datetime:      event.Datetime,
		Seeders:       event.Seeder,
		Leechers:      event.Leechers,
		Interval:      int(event.Interval.Seconds()),
	}

	for i, h := range tr.AnnounceHistory {
		if i == (announceHistoryMaxLength - 1) {
			break
		}
		history[i+1] = h
	}

	tr.AnnounceHistory = history

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onTorrentAnnounceFailed(event broadcast.TorrentAnnounceFailedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	tr := l.state.Torrents[event.Infohash.String()].Trackers[event.TrackerUrl]
	tr.IsAnnouncing = false
	tr.Interval = -1
	tr.Seeders = 0
	tr.Leechers = 0

	history := make([]*AnnounceResult, announceHistoryMaxLength)
	history[0] = &AnnounceResult{
		AnnounceEvent: event.AnnounceEvent,
		WasSuccessful: false,
		Datetime:      event.Datetime,
		Error:         event.Error,
	}

	for i, h := range tr.AnnounceHistory {
		if i == (announceHistoryMaxLength - 1) {
			break
		}
		history[i+1] = h
	}

	tr.AnnounceHistory = history

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onTorrentSwarmChanged(event broadcast.TorrentSwarmChangedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	t := l.state.Torrents[event.Infohash.String()]
	t.Seeders = event.Seeder
	t.Leechers = event.Leechers

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onTorrentRemoved(event broadcast.TorrentRemovedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	delete(l.state.Torrents, event.Infohash.String())

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onNoticeableError(event broadcast.NoticeableErrorEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onGlobalBandwidthChanged(event broadcast.GlobalBandwidthChangedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.state.Bandwidth == nil {
		l.state.Bandwidth = &Bandwidth{
			Torrents: map[string]*TorrentBandwidth{},
		}
	}

	l.state.Bandwidth.CurrentBandwidth = event.AvailableBandwidth

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) onBandwidthWeightHasChanged(event broadcast.BandwidthWeightHasChangedEvent) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.state.Bandwidth == nil {
		l.state.Bandwidth = &Bandwidth{}
	}

	newBandwidthMap := make(map[string]*TorrentBandwidth, len(event.TorrentWeights))

	for infohash, weight := range event.TorrentWeights {
		newBandwidthMap[infohash.String()] = &TorrentBandwidth{
			Infohash:           infohash.String(),
			PercentOfBandwidth: float32(event.TotalWeight) / float32(weight),
		}
	}

	l.state.Bandwidth.Torrents = newBandwidthMap

	// TODO: dispatch to websocket
}

func (l *AppStateCoreListener) hasTorrent(infohash torrent.InfoHash) bool {
	_, exists := l.state.Torrents[infohash.String()]

	return !exists
}
