package seed

//go:generate mockgen -destination=./seed_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/seed -package=seed github.com/anthonyraymond/joal-cli/pkg/seed ISeed

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math"
	"path/filepath"
	"sync"
	"time"
)

type ISeed interface {
	FilePath() string
	InfoHash() *torrent.InfoHash
	TorrentName() string
	GetSwarm() bandwidth.ISwarm
	Seed(bitTorrentClient emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher)
	StopSeeding(ctx context.Context)
}
type seed struct {
	path              string
	torrentSpecs      *torrent.TorrentSpec
	infoHash          *torrent.InfoHash
	announceList      metainfo.AnnounceList
	seeding           bool
	nextAnnounce      tracker.AnnounceEvent
	nextAnnounceAt    time.Time
	seedingStats      *seedStats
	peers             bandwidth.ISwarm
	lastKnownInterval time.Duration
	consecutiveErrors int32
	stop              chan bool     // channel to stop the seed
	stopped           chan struct{} // channel that receives a signal when to seed has been fully terminated
	lock              *sync.Mutex
}

func (s *seed) FilePath() string {
	return s.path
}

func (s *seed) InfoHash() *torrent.InfoHash {
	return s.infoHash
}
func (s *seed) TorrentName() string {
	return s.torrentSpecs.DisplayName
}
func (s *seed) AddUploaded(bytes int64) {
	s.seedingStats.AddUploaded(bytes)
}
func (s *seed) GetSwarm() bandwidth.ISwarm {
	return s.peers
}

func (s *seed) String() string {
	return fmt.Sprintf("%v", &s)
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
		torrentSpecs:      torrent.TorrentSpecFromMetaInfo(info),
		infoHash:          &infoHash,
		announceList:      announceList,
		seeding:           false,
		nextAnnounce:      tracker.Started,
		nextAnnounceAt:    time.Now(),
		seedingStats:      &seedStats{Downloaded: 0, Left: 0, Uploaded: 0},
		peers:             nil,
		lastKnownInterval: 5 * time.Second,
		consecutiveErrors: 0,
		stop:              make(chan bool),
		stopped:           make(chan struct{}),
		lock:              &sync.Mutex{},
	}, nil
}

func (s *seed) Seed(bitTorrentClient emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) {
	s.lock.Lock()
	if s.seeding {
		// TODO: log already running
		s.lock.Unlock()
		return
	}
	logrus.WithFields(logrus.Fields{
		"torrent": filepath.Base(s.path),
	}).Info("Start seed")

	defer func() {
		s.seeding = false
		dispatcher.Release(s)
		close(s.stopped)
	}()

	s.seeding = true
	s.nextAnnounce = tracker.Started
	s.nextAnnounceAt = time.Now()

	s.lock.Unlock()

	for {
		announceAfter := time.NewTimer(time.Until(s.nextAnnounceAt))

		select {
		case <-announceAfter.C:
			currentAnnounceType := s.nextAnnounce
			response, err := bitTorrentClient.Announce(&s.announceList, *s.infoHash, s.seedingStats.Uploaded, s.seedingStats.Downloaded, s.seedingStats.Left, currentAnnounceType)
			if err != nil {
				s.consecutiveErrors = s.consecutiveErrors + 1
				if currentAnnounceType == tracker.None {
					// we already had an interval returned by the tracker, just reuse it
					s.nextAnnounceAt = time.Now().Add(s.lastKnownInterval)
				} else {
					// When error occurs, increment announce time from 10 sec up to 1800 s (
					progressiveDuration := math.Min(1800, float64(10*(s.consecutiveErrors*s.consecutiveErrors)))
					s.nextAnnounceAt = time.Now().Add(time.Duration(progressiveDuration) * time.Second)
				}
				logrus.WithError(err).WithField("infohash", s.infoHash).Warn("failed to announce")
				if s.consecutiveErrors >= 2 && currentAnnounceType != tracker.Started {
					s.peers = &swarm{seeders: 0, leechers: 0}
					dispatcher.ClaimOrUpdate(s)
				}
				continue
			}
			s.consecutiveErrors = 0
			if currentAnnounceType == tracker.Stopped {
				return
			}

			s.lastKnownInterval = time.Duration(response.Interval) * time.Second
			s.nextAnnounce = tracker.None
			s.nextAnnounceAt = time.Now().Add(s.lastKnownInterval)
			// seeders = seeders -1 because we count as one
			s.peers = &swarm{leechers: response.Leechers, seeders: int32(math.Max(0, float64(response.Seeders)-1))}
			dispatcher.ClaimOrUpdate(s)

			continue
		case <-s.stop:
			s.consecutiveErrors = 0
			s.nextAnnounceAt = time.Now()
			s.nextAnnounce = tracker.Stopped

			announceAfter.Stop() // Stop the timer and drain the channel to prevent memory leak
			select {
			case <-announceAfter.C: // if message has arrived concurently drain in and do nothing
			default: // if no message use the default immediatly to exit le select
			}

			continue
		}
	}
}

func (s *seed) StopSeeding(ctx context.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.seeding {
		return
	}

	// prefer sending a value to the channel over closing the channel. Since the select is in a loop, when the channel is closed the select will go automatically to the closed chan case most of the time and will almost never announce stop
	s.stop <- true

	s.seeding = false

	// Wait till context expires or the seed has exited
	select {
	case <-ctx.Done():
		logrus.WithFields(logrus.Fields{
			"torrent": filepath.Base(s.path),
		}).Warn("Seed has not stopped gracefully exiting due to context timeout")
	case <-s.stopped:
		logrus.WithFields(logrus.Fields{
			"torrent": filepath.Base(s.path),
		}).Info("Seed stopped gracefully")
	}
}
