package bandwidth

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/internal/old/core"
	"github.com/anthonyraymond/joal-cli/internal/old/core/broadcast"
	"github.com/anthonyraymond/joal-cli/internal/old/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/old/utils/dataunit"
	"github.com/anthonyraymond/joal-cli/internal/old/utils/stop"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Peers struct {
	Leechers int32
	Seeders  int32
}

type RegisteredTorrent struct {
	InfoHash torrent.InfoHash
	GetPeers func() *Peers
	SetSpeed func(bps int64)
}

type SpeedDispatcher interface {
	Start(config *core.DispatcherConfig)
	Stop()
	ReplaceSpeedConfig(config *core.SpeedProviderConfig)
	Register(rt *RegisteredTorrent) (unregisterTorrent func())
}

type speedDispatcherImpl struct {
	updateTorrentSpeedInterval time.Duration
	randomSpeedProvider        iRandomSpeedProvider
	isRunning                  bool
	stopping                   stop.Chan
	lock                       *sync.Mutex
	torrents                   *registeredTorrentList
}

func NewSpeedDispatcher(conf *core.SpeedProviderConfig) SpeedDispatcher {
	s := &speedDispatcherImpl{
		updateTorrentSpeedInterval: 20 * time.Second,
		randomSpeedProvider:        newRandomSpeedProvider(conf),
		isRunning:                  false,
		stopping:                   stop.NewChan(),
		lock:                       &sync.Mutex{},
		torrents:                   newRegisteredTorrentList(),
	}

	return s
}

func (s *speedDispatcherImpl) Start(config *core.DispatcherConfig) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isRunning {
		return
	}
	s.isRunning = true

	go func(s *speedDispatcherImpl) {
		logger := logs.GetLogger()
		refreshBandwidthTicker := time.NewTicker(config.GlobalBandwidthRefreshInterval)
		updateTorrentSpeedTicker := time.NewTicker(s.updateTorrentSpeedInterval)
		for {
			select {
			case <-refreshBandwidthTicker.C:
				s.randomSpeedProvider.Refresh()
				broadcast.EmitGlobalBandwidthChanged(broadcast.GlobalBandwidthChangedEvent{AvailableBandwidth: s.randomSpeedProvider.GetBytesPerSeconds()})
				logger.Info("bandwidth dispatcher: refreshed global bandwidth",
					zap.String("available-bandwidth", fmt.Sprintf("%s/s", dataunit.ByteCountSI(s.randomSpeedProvider.GetBytesPerSeconds()))),
				)
			case <-updateTorrentSpeedTicker.C:
				updateSpeed(s.torrents.List(), s.randomSpeedProvider.GetBytesPerSeconds())
			case stopRequest := <-s.stopping:
				//goland:noinspection GoDeferInLoop
				defer func() {
					stopRequest.NotifyDone()
				}()

				// close & drain refreshBandwidthTicker
				refreshBandwidthTicker.Stop()
				select {
				case <-refreshBandwidthTicker.C:
				default:
				}
				// close & drain updateTorrentSpeedTicker
				updateTorrentSpeedTicker.Stop()
				select {
				case <-updateTorrentSpeedTicker.C:
				default:
				}

				s.torrents.Reset()

				return
			}
		}
	}(s)
}

func updateSpeed(torrents []*RegisteredTorrent, currentBandwidth int64) {
	if len(torrents) == 0 {
		return
	}

	sumOfLeechers := float64(0)
	for _, registeredTorrent := range torrents {
		p := registeredTorrent.GetPeers()
		leech := float64(p.Leechers)

		sumOfLeechers += leech
	}

	totalWeight := float64(0)
	weights := make([]float64, len(torrents))
	for i, registeredTorrent := range torrents {
		p := registeredTorrent.GetPeers()
		if p.Leechers == 0 || p.Seeders == 0 {
			weights[i] = 0
			continue
		}
		leech := float64(p.Leechers)
		seed := float64(p.Seeders)

		seederRatio := leech / (leech + seed)       // more seeders compared to leecher the better
		leechersPercentage := leech / sumOfLeechers // more seeder compared to total number of seeders the better

		weight := (seederRatio + leechersPercentage) / 2 // sum the two ratio and divide by two to get a number between 0 and 1
		totalWeight += weight
		weights[i] = weight
	}
	for i, registeredTorrent := range torrents {
		p := registeredTorrent.GetPeers()
		if p.Leechers == 0 || p.Seeders == 0 {
			registeredTorrent.SetSpeed(0)
			continue
		}
		percentageOfWeight := weights[i] / totalWeight
		registeredTorrent.SetSpeed(int64(percentageOfWeight * float64(currentBandwidth)))
	}
}

func (s *speedDispatcherImpl) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.isRunning {
		return
	}
	s.isRunning = false

	logger := logs.GetLogger()

	stopReq := stop.NewRequest(context.Background())
	logger.Info("bandwidth dispatcher: stopping")
	s.stopping <- stopReq

	_ = stopReq.AwaitDone()
	logger.Info("bandwidth dispatcher: stopped")
}

func (s *speedDispatcherImpl) ReplaceSpeedConfig(config *core.SpeedProviderConfig) {
	s.randomSpeedProvider.ReplaceSpeedConfig(config)
}

func (s *speedDispatcherImpl) Register(rt *RegisteredTorrent) (unregisterTorrent func()) {
	s.torrents.Add(rt)

	unregisterTorrent = func() {
		s.torrents.Remove(rt)
	}
	return unregisterTorrent
}

type registeredTorrentList struct {
	torrents map[string]*RegisteredTorrent
	lock     *sync.RWMutex
}

func newRegisteredTorrentList() *registeredTorrentList {
	return &registeredTorrentList{
		torrents: make(map[string]*RegisteredTorrent),
		lock:     &sync.RWMutex{},
	}
}

func (l *registeredTorrentList) Add(registeredTorrent *RegisteredTorrent) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.torrents[registeredTorrent.InfoHash.HexString()] = registeredTorrent
}

func (l *registeredTorrentList) List() []*RegisteredTorrent {
	l.lock.RLock()
	defer l.lock.RUnlock()
	v := make([]*RegisteredTorrent, 0, len(l.torrents))

	for _, value := range l.torrents {
		v = append(v, value)
	}
	return v
}

func (l *registeredTorrentList) Remove(registeredTorrent *RegisteredTorrent) {
	l.lock.Lock()
	defer l.lock.Unlock()
	r, exists := l.torrents[registeredTorrent.InfoHash.HexString()]
	if exists {
		r.SetSpeed(0)
		delete(l.torrents, registeredTorrent.InfoHash.HexString())
	}
}

func (l *registeredTorrentList) Reset() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.torrents = make(map[string]*RegisteredTorrent)
}
