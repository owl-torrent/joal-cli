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
	"sync"
	"time"
)

type status int

type ISeed interface {
	FilePath() string
	InfoHash() *torrent.InfoHash
	GetSwarm() bandwidth.ISwarm
	Seed(bitTorrentClient emulatedclients.IEmulatedClient, dispatcher bandwidth.IDispatcher)
	StopSeeding(ctx context.Context)
}
type seed struct {
	path              string
	infoHash          *torrent.InfoHash
	announceList      metainfo.AnnounceList
	seeding           bool
	nextAnnounce      tracker.AnnounceEvent
	nextAnnounceAt    time.Time
	seedingStats      *seedStats
	peers             bandwidth.ISwarm
	lastKnownInterval time.Duration
	consecutiveErrors int32
	stop              chan struct{} // channel to stop the seed
	stopped           chan struct{} // channel that receives a signal when to seed has been fully terminated
	lock              *sync.Mutex
}

func (t *seed) FilePath() string {
	return t.path
}

func (t *seed) InfoHash() *torrent.InfoHash {
	return t.infoHash
}
func (t *seed) AddUploaded(bytes int64) {
	t.seedingStats.AddUploaded(bytes)
}
func (t *seed) GetSwarm() bandwidth.ISwarm {
	return t.peers
}

func LoadFromFile(file string) (ISeed, error) {
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
	return &seed{
		path:              file,
		infoHash:          &infoHash,
		announceList:      announceList,
		seeding:           false,
		nextAnnounce:      tracker.Started,
		nextAnnounceAt:    time.Now(),
		seedingStats:      &seedStats{Downloaded: 0, Left: 0, Uploaded: 0},
		peers:             nil,
		lastKnownInterval: 5 * time.Second,
		consecutiveErrors: 0,
		stop:              make(chan struct{}),
		stopped:           make(chan struct{}),
		lock:              &sync.Mutex{},
	}, nil
}

func (t *seed) Seed(bitTorrentClient emulatedclients.IEmulatedClient, dispatcher bandwidth.IDispatcher) {
	t.lock.Lock()
	if t.seeding {
		// TODO: log already running
		t.lock.Unlock()
		return
	}

	defer func() {
		t.seeding = false
		close(t.stopped)
	}()

	t.seeding = true
	t.nextAnnounce = tracker.Started
	t.nextAnnounceAt = time.Now()

	t.lock.Unlock()

	for {
		announceAfter := time.NewTimer(time.Until(t.nextAnnounceAt))

		select {
		case <-announceAfter.C:
			currentAnnounceType := t.nextAnnounce
			response, err := bitTorrentClient.Announce(&t.announceList, *t.infoHash, t.seedingStats.Uploaded, t.seedingStats.Downloaded, t.seedingStats.Left, currentAnnounceType)
			if err != nil {
				t.consecutiveErrors = t.consecutiveErrors + 1
				if currentAnnounceType == tracker.None {
					// we already had an interval returned by the tracker, just reuse it
					t.nextAnnounceAt = time.Now().Add(t.lastKnownInterval)
				} else {
					// increment announce time from 10 sec up to 1800 s (
					progressiveDuration := math.Min(1800, float64(10*(t.consecutiveErrors*t.consecutiveErrors)))
					t.nextAnnounceAt = time.Now().Add(time.Duration(progressiveDuration) * time.Second)
				}
				// TODO: log announce error
				if t.consecutiveErrors > 2 && currentAnnounceType != tracker.Started {
					t.peers = &swarm{seeders: 0, leechers: 0}
				}
				dispatcher.ClaimOrUpdate(t)
				continue
			}
			t.consecutiveErrors = 0
			if currentAnnounceType == tracker.Stopped {
				dispatcher.Release(t)
				return
			}

			t.lastKnownInterval = time.Duration(response.Interval) * time.Second
			t.nextAnnounce = tracker.None
			t.nextAnnounceAt = time.Now().Add(t.lastKnownInterval)
			t.peers = &swarm{leechers: response.Leechers, seeders: response.Seeders}
			dispatcher.ClaimOrUpdate(t)

			continue
		case <-t.stop:
			t.consecutiveErrors = 0
			t.nextAnnounceAt = time.Now()
			t.nextAnnounce = tracker.Stopped

			announceAfter.Stop() // Stop the timer and drain the channel to prevent memory leak
			select {
			case <-announceAfter.C: // if message has arrived concurently drain in and do nothing
			default: // if no message use the default immediatly to exit le select
			}
			continue
		}
	}
}

func (t *seed) StopSeeding(ctx context.Context) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.seeding {
		return
	}

	close(t.stop)
	t.seeding = false

	// Wait till context expires or the seed has exited
	select {
	case <-ctx.Done():
		//TODO: log return by timeout
	case <-t.stopped:
		//TODO: log gracefully shutted down
	}
}
