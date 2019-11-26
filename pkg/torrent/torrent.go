package torrent

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients"
	"github.com/pkg/errors"
	"time"
)

type status int

const (
	onHold status = iota
	seeding
)

type Torrent struct {
	infoHash          *torrent.InfoHash
	announceList      metainfo.AnnounceList
	currentStatus     status
	nextAnnounce      tracker.AnnounceEvent
	nextAnnounceAt    time.Time
	seedingStats      *seedStats
	peers             bandwidth.ISwarm
	bitTorrentClient  *emulatedclients.EmulatedClient
	lastKnownInterval time.Duration
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

func LoadFromFile(file string, bitTorrentClient *emulatedclients.EmulatedClient) (*Torrent, error) {
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
		infoHash:          &infoHash,
		announceList:      announceList,
		currentStatus:     onHold,
		nextAnnounce:      tracker.Started,
		nextAnnounceAt:    time.Now(),
		seedingStats:      &seedStats{Downloaded: 0, Left: 0, Uploaded: 0},
		peers:             nil,
		bitTorrentClient:  bitTorrentClient,
		lastKnownInterval: 5 * time.Second,
	}, nil
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
		for ; ; {
			announceAfter := time.After(time.Until(t.nextAnnounceAt))

			select {
			case <-announceAfter:
				response, err := t.bitTorrentClient.Announce(&t.announceList, *t.infoHash, t.seedingStats.Uploaded, t.seedingStats.Downloaded, t.seedingStats.Left, t.nextAnnounce)
				if err != nil {
					if t.nextAnnounce != tracker.None {
						t.nextAnnounceAt = time.Now().Add(t.lastKnownInterval)
					} else {
						t.nextAnnounceAt = time.Now().Add(10 * time.Second)
					}
					// TODO: log announce error
					break
				}
				if t.nextAnnounce == tracker.Stopped {
					return
				}

				t.lastKnownInterval = time.Duration(response.Interval) * time.Second
				t.nextAnnounce = tracker.None

				return
				/*case <-stopGracefull:
					return
				case <-killNow:
					return*/
			}
		}
	}(t)
}

func (t *Torrent) StopSeeding() {
	if t.currentStatus != seeding {
		return
	}

}
