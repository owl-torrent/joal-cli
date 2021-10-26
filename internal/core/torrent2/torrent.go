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
	path      string
	infoHash  torrent.InfoHash
	stats     Stats
	peers     Peers
	trackers  []*trackerImpl
	metaInfo  *slimMetaInfo
	info      *slimInfo
	isRunning bool
	stopping  stop.Chan
	lock      *sync.Mutex
}

func FromFile(filePath string) (Torrent, error) {
	log := logs.GetLogger().With(zap.String("torrent", filepath.Base(filePath)))
	meta, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load meta-info from file '%s'", filePath)
	}

	info, err := meta.UnmarshalInfo()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load info from file '%s'", filePath)
	}
	infoHash := meta.HashInfoBytes()
	log.Info("torrent: parsed successfully", zap.ByteString("infohash", infoHash.Bytes()))

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

	log := logs.GetLogger().With(zap.String("torrent", filepath.Base(t.path)))

	stopReq := stop.NewRequest(ctx)
	log.Info("torrent: stopping")
	t.stopping <- stopReq

	_ = stopReq.AwaitDone()
	log.Info("torrent: stopped")
}

func torrentRoutine(t *torrentImpl, props AnnounceProps) {
	t.peers.Reset()
	t.stats.Reset()

	onAnnounceSuccess := make(chan emulatedclient.AnnounceResponse, len(t.trackers))
	onAnnounceError := make(chan emulatedclient.AnnounceResponseError, len(t.trackers))
	onAnnounceTime := time.After(0 * time.Second)

	announceCallbacks := &emulatedclient.AnnounceCallbacks{
		Success: func(response emulatedclient.AnnounceResponse) {
			if response.Request.Event == tracker.Stopped {
				return
			}
			onAnnounceSuccess <- response
		},
		Failed: func(responseError emulatedclient.AnnounceResponseError) {
			if responseError.Request.Event == tracker.Stopped {
				return
			}
			onAnnounceError <- responseError
		},
	}

	for {
		select {
		case resp := <-onAnnounceSuccess:
			currentTracker := findTracker(resp.Request.Url, t.trackers)
			if currentTracker != nil {
				currentTracker.state.startSent = true
				currentTracker.Succeed(AnnounceHistory{
					interval: resp.Interval,
					seeders:  resp.Seeders,
					leechers: resp.Leechers,
				})
			}

			//TODO prioritize tracker in his tier

			onAnnounceTime = getNextAnnounceTime()

		case errorResponse := <-onAnnounceError:
			currentTracker := findTracker(errorResponse.Request.Url, t.trackers)
			if currentTracker != nil {
				currentTracker.Succeed(AnnounceHistory{
					error: errorResponse.Error(),
				})
			}
			// TODO: handle de-prioritization of the tracker

			onAnnounceTime = getNextAnnounceTime(t.trackers)
		case <-onAnnounceTime:
			t.announceToTrackers(props, announceCallbacks, tracker.None)

		case stopRequest := <-t.stopping:
			//goland:noinspection GoDeferInLoop
			defer func() {
				stopRequest.NotifyDone()
			}()
			if stopRequest.Ctx().Err() != nil {
				// context is already expired, no need to announce stop
				return
			}
			t.announceToTrackers(props, announceCallbacks, tracker.Stopped)

			return
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

// announceAbleTrackers return all the tracker able to announce at the moment
func findAnnounceReadyTrackers(trackers []*trackerImpl, announceToAllTier bool, announceToAllTracker bool) []*trackerImpl {
	var announceAbleTrackers []*trackerImpl

	now := time.Now()
	// index of the tier we last found and announceable tracker
	__foundForTier := int16(-1)
	__foundOne := false

	for i, tr := range trackers {
		if !tr.enabled {
			continue
		}
		if announceToAllTier && !announceToAllTracker && __foundForTier == tr.tier {
			continue
		}
		// Announcing to a single tracker in a single tier => we found one => exit
		if !announceToAllTier && !announceToAllTracker && __foundOne {
			return announceAbleTrackers
		}
		// Announcing to all trackers in one tier => we have found at least one and changed tier => exit
		if !announceToAllTier && announceToAllTracker && __foundOne && i > 0 && tr.tier > trackers[i-1].tier {
			return announceAbleTrackers
		}

		if tr.CanAnnounce(now) {
			__foundOne = true
			__foundForTier = tr.tier
			announceAbleTrackers = append(announceAbleTrackers, tr)
			continue
		}
		// Can not announce ATM but the tracker is working, flag that we found trackers
		if tr.IsWorking() {
			__foundOne = true
			__foundForTier = tr.tier
		}
	}
	return announceAbleTrackers
}

func findTracker(u url.URL, trackers []*trackerImpl) *trackerImpl {
	u.RawQuery = ""
	u.RawFragment = ""
	for _, t := range trackers {
		currentUrl := t.Url()
		if strings.EqualFold(currentUrl.String(), u.String()) {
			return t
		}
	}

	return nil
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
