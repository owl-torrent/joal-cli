package torrent2

import (
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/utils/stop"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var randSeed = time.Now().UnixNano()

type Torrent interface {
}

type torrentImpl struct {
	path     string
	infoHash torrent.InfoHash
	stats    Stats
	peers    Peers
	trackers []*trackerImpl
	metaInfo *slimMetaInfo
	info     *slimInfo

	isRunning bool
	stopping  stop.Chan
	lock      *sync.Mutex
}

func FromFile(filePath string) (Torrent, error) {
	logger := logs.GetLogger().With(zap.String("torrent", filepath.Base(filePath)))
	meta, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load meta-info from file '%s'", filePath)
	}

	info, err := meta.UnmarshalInfo()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load info from file '%s'", filePath)
	}
	infoHash := meta.HashInfoBytes()
	logger.Info("torrent: parsed successfully", zap.ByteString("infohash", infoHash.Bytes()))

	private := false
	if info.Private != nil && *info.Private == true {
		private = true
	}

	return &torrentImpl{
		path:  filePath,
		stats: newStats(),
		peers: newPeersElector(),
		metaInfo: &slimMetaInfo{
			Announce:     meta.Announce,
			AnnounceList: meta.AnnounceList,
			Nodes:        meta.Nodes,
			CreationDate: meta.CreationDate,
			Comment:      meta.Comment,
			CreatedBy:    meta.CreatedBy,
			Encoding:     meta.Encoding,
			UrlList:      meta.UrlList,
		},
		info: &slimInfo{
			PieceLength: info.PieceLength,
			Name:        info.Name,
			Length:      info.Length,
			Private:     private,
			Source:      info.Source,
		},
		trackers:  []*trackerImpl{},
		infoHash:  infoHash,
		isRunning: false,
		stopping:  stop.NewChan(),
		lock:      &sync.Mutex{},
	}, nil
}

type AnnounceProps struct {
	SupportAnnounceList   bool
	SupportHttpAnnounce   bool
	SupportUdpAnnounce    bool
	AnnounceToAllTiers    bool
	AnnounceToAllTrackers bool
}

func (t *torrentImpl) Start(props AnnounceProps) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.isRunning {
		return
	}
	t.isRunning = true

	t.trackers = newTrackers(t.metaInfo.Announce, t.metaInfo.AnnounceList, props.SupportAnnounceList)

	// Disable trackers based on client capabilities (UDP, HTTP, ...)
	for _, track := range t.trackers {
		if strings.Contains(strings.ToLower(track.Url().Scheme), "http") && !props.SupportHttpAnnounce {
			track.enabled = false
		} else if strings.Contains(strings.ToLower(track.Url().Scheme), "udp") && !props.SupportUdpAnnounce {
			track.enabled = false
		}
	}

	go torrentRoutine(t, props)
}

func (t *torrentImpl) Stop(ctx context.Context) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.isRunning {
		return
	}
	t.isRunning = false

	logger := logs.GetLogger().With(zap.String("torrent", filepath.Base(t.path)))

	stopReq := stop.NewRequest(ctx)
	logger.Info("torrent: stopping")
	t.stopping <- stopReq

	_ = stopReq.AwaitDone()
	logger.Info("torrent: stopped")
}

func torrentRoutine(t *torrentImpl, props AnnounceProps) {
	logger := logs.GetLogger().With(zap.String("torrent", t.info.Name))
	t.peers.Reset()
	t.stats.Reset()

	onAnnounceSuccess := make(chan emulatedclient.AnnounceResponse, len(t.trackers))
	onAnnounceError := make(chan emulatedclient.AnnounceResponseError, len(t.trackers))

	timer := time.NewTimer(0 * time.Second)
	onAnnounceTime := timer.C

	dismissAnnounceResults := atomic.NewBool(false)
	announceCallbacks := &emulatedclient.AnnounceCallbacks{
		Success: func(response emulatedclient.AnnounceResponse) {
			if dismissAnnounceResults.Load() {
				return
			}
			onAnnounceSuccess <- response
		},
		Failed: func(responseError emulatedclient.AnnounceResponseError) {
			if dismissAnnounceResults.Load() {
				return
			}
			onAnnounceError <- responseError
		},
	}

	for {
		select {
		case resp := <-onAnnounceSuccess:
			_, currentTracker := findTracker(resp.Request.Url, t.trackers)
			if currentTracker != nil {
				currentTracker.state.startSent = true
				currentTracker.Succeed(AnnounceHistory{
					interval: resp.Interval,
					seeders:  resp.Seeders,
					leechers: resp.Leechers,
				})
			}

			t.peers.AddPeer(SwarmUpdateRequest{
				trackerUrl: currentTracker.Url(),
				interval:   resp.Interval,
				seeders:    resp.Seeders,
				leechers:   resp.Leechers,
			})

			if !timer.Stop() {
				<-timer.C
			}
			nextAnnounce := getNextAnnounceTime(t.trackers, props.AnnounceToAllTiers, props.AnnounceToAllTrackers)
			if nextAnnounce.IsZero() {
				logger.Error("getNextAnnounceTime returned a 0 time, this should not happen since the function should only be called after an announce is done. Thus there should always be a tracker to announce next")
				onAnnounceTime = nil
			} else {
				timer = time.NewTimer(nextAnnounce.Sub(time.Now()))
				onAnnounceTime = timer.C
			}

		case errorResponse := <-onAnnounceError:
			trackerIndex, currentTracker := findTracker(errorResponse.Request.Url, t.trackers)
			if currentTracker != nil {
				currentTracker.Failed(AnnounceHistory{
					error: errorResponse.Error(),
				}, 250, int(errorResponse.Interval.Seconds()))
			}
			t.peers.AddPeer(SwarmUpdateRequest{
				trackerUrl: currentTracker.Url(),
				interval:   0, // set interval to 0 will force the entry to be evicted by the peer electors system
				seeders:    0,
				leechers:   0,
			})
			deprioritizeTracker(t.trackers, trackerIndex)

			if !timer.Stop() {
				<-timer.C
			}
			nextAnnounce := getNextAnnounceTime(t.trackers, props.AnnounceToAllTiers, props.AnnounceToAllTrackers)
			if nextAnnounce.IsZero() {
				logger.Error("getNextAnnounceTime returned a 0 time, this should not happen since the function should only be called after an announce is done. Thus there should always be a tracker to announce next")
				onAnnounceTime = nil
			} else {
				timer = time.NewTimer(nextAnnounce.Sub(time.Now()))
				onAnnounceTime = timer.C
			}
		case <-onAnnounceTime:
			t.announceToTrackers(props, announceCallbacks, tracker.None)

		case stopRequest := <-t.stopping:
			//goland:noinspection GoDeferInLoop
			defer func() {
				stopRequest.NotifyDone()
			}()
			dismissAnnounceResults.Store(true)
			//drain announce response channels
			drainSuccessResponseChan(onAnnounceSuccess)
			drainErrorResponseChan(onAnnounceError)

			if stopRequest.Ctx().Err() != nil {
				// context is already expired, no need to announce stop
				return
			}
			for _, tr := range t.trackers {
				tr.state.nextAnnounce = time.Now()
			}
			t.announceToTrackers(props, announceCallbacks, tracker.Stopped)

			return
		}
	}
}

/**
deprioritizeTracker push a tracker to the end of his tier
*/
func deprioritizeTracker(trackers []*trackerImpl, indexToDeprioritize int) {
	if indexToDeprioritize >= len(trackers)-1 {
		// out of bound or already the last one
		return
	}
	trackerToDeprioritize := trackers[indexToDeprioritize]

	for i := indexToDeprioritize; i < len(trackers); i++ {
		if i+1 == len(trackers) {
			return
		}
		t := trackers[i]
		if t.tier > trackerToDeprioritize.tier {
			return
		}

		if trackers[i+1].tier == trackerToDeprioritize.tier {
			// swap
			trackers[i], trackers[i+1] = trackers[i+1], trackers[i]
		}
	}
}

func (t *torrentImpl) announceToTrackers(props AnnounceProps, callbacks *emulatedclient.AnnounceCallbacks, event tracker.AnnounceEvent) {
	trackersToAnnounce := findAnnounceReadyTrackers(t.trackers, props.AnnounceToAllTiers, props.AnnounceToAllTrackers)

	for _, currentTracker := range trackersToAnnounce {
		if event == tracker.Stopped && !currentTracker.state.startSent {
			// no need to announce stop of not started
			continue
		}
		currentTracker.state.updating = true
		if event == tracker.None && !currentTracker.state.startSent {
			event = tracker.Started
		}
		req := emulatedclient.AnnounceRequest{
			Url:               currentTracker.Url(),
			InfoHash:          t.infoHash,
			Downloaded:        t.stats.Downloaded(),
			Left:              t.stats.Left(),
			Uploaded:          t.stats.Uploaded(),
			Corrupt:           t.stats.Corrupted(),
			Event:             event,
			Private:           t.info.Private,
			AnnounceCallbacks: callbacks,
		}

		queueAnnounce(req)
	}
}

func getNextAnnounceTime(trackers []*trackerImpl, announceToAllTier bool, announceToAllTracker bool) time.Time {
	nextAnnounce := time.Time{}

	foundForTier := int16(-1)
	foundOne := false

	for i, tr := range trackers {
		if !tr.enabled {
			continue
		}
		if announceToAllTier && !announceToAllTracker && foundForTier == tr.tier {
			continue
		}
		// Announcing to a single tracker in a single tier => we found one => exit
		if !announceToAllTier && !announceToAllTracker && foundOne {
			return nextAnnounce
		}
		// Announcing to all trackers in one tier => we have found at least one and changed tier => exit
		if !announceToAllTier && announceToAllTracker && foundOne && i > 0 && tr.tier > trackers[i-1].tier {
			return nextAnnounce
		}

		// set flags to instruct we found a tracker in this tier and a working tracker
		if tr.state.updating {
			foundForTier = tr.tier
			foundOne = true
			continue
		}

		if nextAnnounce.IsZero() || nextAnnounce.After(tr.state.nextAnnounce) {
			nextAnnounce = tr.state.nextAnnounce
		}
	}

	return nextAnnounce
}

// announceAbleTrackers return all the tracker able to announce at the moment
func findAnnounceReadyTrackers(trackers []*trackerImpl, announceToAllTier bool, announceToAllTracker bool) []*trackerImpl {
	var announceAbleTrackers []*trackerImpl

	now := time.Now()
	// index of the tier we last found and announce-ready tracker in
	foundForTier := int16(-1)
	foundOne := false

	for i, tr := range trackers {
		if !tr.enabled {
			continue
		}
		if announceToAllTier && !announceToAllTracker && foundForTier == tr.tier {
			continue
		}
		// Announcing to a single tracker in a single tier => we found one => exit
		if !announceToAllTier && !announceToAllTracker && foundOne {
			return announceAbleTrackers
		}
		// Announcing to all trackers in one tier => we have found at least one and changed tier => exit
		if !announceToAllTier && announceToAllTracker && foundOne && i > 0 && tr.tier > trackers[i-1].tier {
			return announceAbleTrackers
		}

		if tr.CanAnnounce(now) {
			foundOne = true
			foundForTier = tr.tier
			announceAbleTrackers = append(announceAbleTrackers, tr)
			continue
		}
		// Can not announce ATM but the tracker is working, flag that we found trackers
		if tr.IsWorking() {
			foundOne = true
			foundForTier = tr.tier
		}
	}
	return announceAbleTrackers
}

func findTracker(u url.URL, trackers []*trackerImpl) (int, *trackerImpl) {
	u.RawQuery = ""
	u.RawFragment = ""
	for i, t := range trackers {
		currentUrl := t.Url()
		if strings.EqualFold(currentUrl.String(), u.String()) {
			return i, t
		}
	}

	return 0, nil
}

func drainSuccessResponseChan(c chan emulatedclient.AnnounceResponse) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
func drainErrorResponseChan(c chan emulatedclient.AnnounceResponseError) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}

// metainfo.MetaInfo is RAM consuming because of the size of the Piece[], create our own struct that ignore this field
type slimMetaInfo struct {
	Announce     string
	AnnounceList metainfo.AnnounceList
	Nodes        []metainfo.Node
	CreationDate int64
	Comment      string
	CreatedBy    string
	Encoding     string
	UrlList      metainfo.UrlList
}

// metainfo.Info is RAM consuming because of the size of the Piece[], create our own struct that ignore this field
type slimInfo struct {
	PieceLength int64
	Name        string
	Length      int64
	Private     bool
	Source      string
}
