package web

import (
	"encoding/json"
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/internal/core/broadcast"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/go-stomp/stomp"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/url"
	"sync"
)

const announceHistoryMaxLength = 5

type appStateCoreListener struct {
	state          *State
	lock           *sync.Mutex
	stompPublisher *stomp.Conn
}

func (l *appStateCoreListener) OnSeedStart(event broadcast.SeedStartedEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	l.state.Started = true
	l.state.Client = &Client{
		Name:    event.Client,
		Version: event.Version,
	}

	payload := make(map[string]interface{}, 2)
	payload["started"] = l.state.Started
	payload["client"] = l.state.Client

	err := sendToStompTopic(l.stompPublisher, "/seed/started", payload)
	if err != nil {
		log.Error("Failed to send onSeedStart stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnSeedStop(_ broadcast.SeedStoppedEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	l.state = &State{
		Started:   false,
		Client:    nil,
		Config:    nil,
		Torrents:  nil,
		Bandwidth: nil,
	}

	payload := make(map[string]interface{}, 2)
	payload["started"] = l.state.Started

	err := sendToStompTopic(l.stompPublisher, "/seed/stopped", payload)
	if err != nil {
		log.Error("Failed to send onSeedStop stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnConfigChanged(event broadcast.ConfigChangedEvent) {
	log := logs.GetLogger()
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

	err := sendToStompTopic(l.stompPublisher, "/config/changed", l.state.Config)
	if err != nil {
		log.Error("Failed to send onConfigChanged stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnTorrentAdded(event broadcast.TorrentAddedEvent) {
	log := logs.GetLogger()
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
		normalizedUrl := normalizeTrackerAnnounceUrl(u)
		t.Trackers[normalizedUrl.String()] = &TorrentTrackers{
			Url:             normalizedUrl,
			IsAnnouncing:    false,
			InUse:           false,
			AnnounceHistory: []*AnnounceResult{},
		}
	}

	l.state.Torrents[event.Infohash.String()] = t

	err := sendToStompTopic(l.stompPublisher, "/torrents/new", t)
	if err != nil {
		log.Error("Failed to send onTorrentAdded stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnTorrentAnnouncing(event broadcast.TorrentAnnouncingEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	l.state.Torrents[event.Infohash.String()].Uploaded = event.Uploaded

	tr := l.state.Torrents[event.Infohash.String()].Trackers[normalizeTrackerAnnounceUrl(event.TrackerUrl).String()]
	tr.InUse = true
	tr.IsAnnouncing = true

	err := sendToStompTopic(l.stompPublisher, "/torrents/changed", l.state.Torrents[event.Infohash.String()])
	if err != nil {
		log.Error("Failed to send onTorrentAnnouncing stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnTorrentAnnounceSuccess(event broadcast.TorrentAnnounceSuccessEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	tr := l.state.Torrents[event.Infohash.String()].Trackers[normalizeTrackerAnnounceUrl(event.TrackerUrl).String()]
	tr.IsAnnouncing = false
	tr.Interval = int(event.Interval.Seconds())
	tr.Seeders = event.Seeder
	tr.Leechers = event.Leechers

	newLength := announceHistoryMaxLength
	if len(tr.AnnounceHistory) < newLength {
		newLength = len(tr.AnnounceHistory) + 1
	}
	history := make([]*AnnounceResult, newLength)
	history[0] = &AnnounceResult{
		AnnounceEvent: event.AnnounceEvent.String(),
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

	err := sendToStompTopic(l.stompPublisher, "/torrents/changed", l.state.Torrents[event.Infohash.String()])
	if err != nil {
		log.Error("Failed to send onTorrentAnnounceSuccess stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnTorrentAnnounceFailed(event broadcast.TorrentAnnounceFailedEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	tr := l.state.Torrents[event.Infohash.String()].Trackers[normalizeTrackerAnnounceUrl(event.TrackerUrl).String()]
	tr.IsAnnouncing = false
	tr.Interval = -1
	tr.Seeders = 0
	tr.Leechers = 0

	newLength := announceHistoryMaxLength
	if len(tr.AnnounceHistory) < newLength {
		newLength = len(tr.AnnounceHistory) + 1
	}
	history := make([]*AnnounceResult, newLength)
	history[0] = &AnnounceResult{
		AnnounceEvent: event.AnnounceEvent.String(),
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

	err := sendToStompTopic(l.stompPublisher, "/torrents/changed", l.state.Torrents[event.Infohash.String()])
	if err != nil {
		log.Error("Failed to send onTorrentAnnounceFailed stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnTorrentSwarmChanged(event broadcast.TorrentSwarmChangedEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	t := l.state.Torrents[event.Infohash.String()]
	t.Seeders = event.Seeder
	t.Leechers = event.Leechers

	err := sendToStompTopic(l.stompPublisher, "/torrent/changed", t)
	if err != nil {
		log.Error("Failed to send onTorrentSwarmChanged stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnTorrentRemoved(event broadcast.TorrentRemovedEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.hasTorrent(event.Infohash) {
		return
	}

	delete(l.state.Torrents, event.Infohash.String())

	payload := map[string]string{}
	payload["infohash"] = event.Infohash.String()

	err := sendToStompTopic(l.stompPublisher, "/torrent/removed", payload)
	if err != nil {
		log.Error("Failed to send onTorrentRemoved stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnNoticeableError(event broadcast.NoticeableErrorEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	payload := map[string]interface{}{}
	payload["error"] = event.Error
	payload["datetime"] = event.Datetime

	err := sendToStompTopic(l.stompPublisher, "/error", payload)
	if err != nil {
		log.Error("Failed to send onNoticeableError stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnGlobalBandwidthChanged(event broadcast.GlobalBandwidthChangedEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.state.Bandwidth == nil {
		l.state.Bandwidth = &Bandwidth{
			Torrents: map[string]*TorrentBandwidth{},
		}
	}

	l.state.Bandwidth.CurrentBandwidth = event.AvailableBandwidth

	payload := map[string]interface{}{}
	payload["currentBandwidth"] = l.state.Bandwidth.CurrentBandwidth

	err := sendToStompTopic(l.stompPublisher, "/bandwidth/global-speed/changed", payload)
	if err != nil {
		log.Error("Failed to send onGlobalBandwidthChanged stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) OnBandwidthWeightHasChanged(event broadcast.BandwidthWeightHasChangedEvent) {
	log := logs.GetLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.state.Bandwidth == nil {
		l.state.Bandwidth = &Bandwidth{}
	}

	newBandwidthMap := make(map[string]*TorrentBandwidth, len(event.TorrentWeights))

	for infohash, weight := range event.TorrentWeights {
		newBandwidthMap[infohash.String()] = &TorrentBandwidth{
			Infohash:           infohash.String(),
			PercentOfBandwidth: float32(weight) / float32(event.TotalWeight),
		}
	}

	l.state.Bandwidth.Torrents = newBandwidthMap

	err := sendToStompTopic(l.stompPublisher, "/bandwidth/weights", l.state.Bandwidth.Torrents)
	if err != nil {
		log.Error("Failed to send onBandwidthWeightHasChanged stomp message", zap.Error(err))
	}
}

func (l *appStateCoreListener) hasTorrent(infohash torrent.InfoHash) bool {
	_, exists := l.state.Torrents[infohash.String()]

	return exists
}

func sendToStompTopic(stompPublisher *stomp.Conn, destination string, content interface{}) error {
	body, err := json.Marshal(content)
	if err != nil {
		return errors.Wrap(err, "failed to marshal stomp payload as json")
	}

	err = stompPublisher.Send(destination, "application/json", body)
	if err != nil {
		return errors.Wrapf(err, "failed to send a stomp message to the local server for topic %s", destination)
	}
	return nil
}

func normalizeTrackerAnnounceUrl(u url.URL) *url.URL {
	return &url.URL{
		Scheme:     u.Scheme,
		Host:       u.Host,
		ForceQuery: false,
	}
}
