package seed

import (
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients"
	"github.com/pkg/errors"
	"math"
	"time"
)

type status int

const (
	onHold status = iota
	seeding
)

type OnStopHook func()
type Torrent struct {
	infoHash            *torrent.InfoHash
	announceList        metainfo.AnnounceList
	currentStatus       status
	nextAnnounce        tracker.AnnounceEvent
	nextAnnounceAt      time.Time
	seedingStats        *seedStats
	peers               bandwidth.ISwarm
	bitTorrentClient    *emulatedclients.EmulatedClient
	bandwidthDispatcher bandwidth.IDispatcher
	lastKnownInterval   time.Duration
	consecutiveErrors   int32
	onStopHook          OnStopHook
}

func (t *Torrent) InfoHash() *torrent.InfoHash {
	return t.infoHash
}
func (t *Torrent) AddUploaded(bytes int64) {
	t.seedingStats.AddUploaded(bytes)
}
func (t *Torrent) GetSwarm() bandwidth.ISwarm {
	return t.peers
}

func LoadFromFile(file string, bitTorrentClient *emulatedclients.EmulatedClient, dispatcher bandwidth.IDispatcher) (*Torrent, error) {
	info, err := metainfo.LoadFromFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load torrent file: '%s'", file)
	}

	infoHash := info.HashInfoBytes()
	announceList := info.AnnounceList
	if info.Announce != "" {
		firstTier := make([][]string, 1)
		firstTier[0] = []string{info.Announce}
		announceList = append(firstTier, announceList...)
	}
	return &Torrent{
		infoHash:            &infoHash,
		announceList:        announceList,
		currentStatus:       onHold,
		nextAnnounce:        tracker.Started,
		nextAnnounceAt:      time.Now(),
		seedingStats:        &seedStats{Downloaded: 0, Left: 0, Uploaded: 0},
		peers:               nil,
		bitTorrentClient:    bitTorrentClient,
		bandwidthDispatcher: dispatcher,
		lastKnownInterval:   5 * time.Second,
		consecutiveErrors:   0,
	}, nil
}

func (t *Torrent) WithHook(hook OnStopHook) {
	t.onStopHook = hook
}

func (t *Torrent) Seed() {
	if t.currentStatus == seeding {
		// TODO: log already running
		return
	}
	t.currentStatus = seeding
	t.nextAnnounce = tracker.Started
	t.nextAnnounceAt = time.Now()

	go func(t *Torrent) {
		defer func() {
			if t.onStopHook != nil {
				t.onStopHook()
			}
		}()
		for {
			announceAfter := time.After(time.Until(t.nextAnnounceAt))

			select {
			case <-announceAfter:
				currentAnnounceType := t.nextAnnounce
				response, err := t.bitTorrentClient.Announce(&t.announceList, *t.infoHash, t.seedingStats.Uploaded, t.seedingStats.Downloaded, t.seedingStats.Left, currentAnnounceType)
				if err != nil {
					t.consecutiveErrors = t.consecutiveErrors + 1
					if currentAnnounceType == tracker.None {
						t.nextAnnounceAt = time.Now().Add(t.lastKnownInterval)
					} else {
						// increment announce time from 10 sec up to 1800 s
						progressiveDuration := math.Min(1800, float64(10*(t.consecutiveErrors*t.consecutiveErrors)))
						t.nextAnnounceAt = time.Now().Add(time.Duration(progressiveDuration) * time.Second)
					}
					// TODO: log announce error
					if t.consecutiveErrors > 2 && currentAnnounceType != tracker.Started {
						t.peers = &swarm{seeders: 0, leechers: 0}
					}
					t.bandwidthDispatcher.ClaimOrUpdate(t)
					break
				}
				t.consecutiveErrors = 0
				if currentAnnounceType == tracker.Stopped {
					t.bandwidthDispatcher.Release(t)
					return
				}

				t.lastKnownInterval = time.Duration(response.Interval) * time.Second
				t.nextAnnounce = tracker.None
				t.peers = &swarm{leechers: response.Leechers, seeders: response.Seeders}
				t.bandwidthDispatcher.ClaimOrUpdate(t)

				return
				/*case <-stopGracefull:
					return
				case <-killNow:
					return*/
			}
		}
	}(t)
}

func (t *Torrent) StopSeeding(ctx context.Context) {
	if t.currentStatus != seeding {
		return
	}

	// TODO: implement StopSeeding with channel
}
