package seed

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/announcer"
	"github.com/anthonyraymond/joal-cli/internal/core/bandwidth"
	"github.com/anthonyraymond/joal-cli/internal/core/broadcast"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/core/orchestrator"
	"github.com/anthonyraymond/joal-cli/internal/utils/dataunit"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/rand"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var randSeed = time.Now().UnixNano()

type ITorrent interface {
	InfoHash() torrent.InfoHash
	Name() string
	File() string
	TrackerAnnounceUrls() []url.URL
	Size() int64
	StartSeeding(client emulatedclient.IEmulatedClient, bandwidthClaimerPool bandwidth.IBandwidthClaimerPool) error
	StopSeeding(ctx context.Context)
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
	Private     *bool
	Source      string
	Files       []metainfo.FileInfo
}

type stoppingRequest struct {
	ctx          context.Context
	doneStopping chan struct{}
}

type joalTorrent struct {
	path      string
	metaInfo  *slimMetaInfo
	info      *slimInfo
	infoHash  torrent.InfoHash
	isRunning bool
	stopping  chan *stoppingRequest
	lock      *sync.Mutex
}

func FromFile(filePath string) (ITorrent, error) {
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

	// Shuffling trackers according to BEP-12: https://www.bittorrent.org/beps/bep_0012.html
	rand.Seed(randSeed)
	for _, tier := range meta.AnnounceList {
		rand.Shuffle(len(tier), func(i, j int) {
			tier[i], tier[j] = tier[j], tier[i]
		})
	}

	return &joalTorrent{
		path: filePath,
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
			Private:     info.Private,
			Source:      info.Source,
			Files:       info.Files,
		},
		infoHash:  infoHash,
		isRunning: false,
		stopping:  make(chan *stoppingRequest),
		lock:      &sync.Mutex{},
	}, nil
}

func (t *joalTorrent) InfoHash() torrent.InfoHash {
	return t.infoHash
}

func (t *joalTorrent) Name() string {
	return t.info.Name
}

func (t *joalTorrent) File() string {
	return t.path
}
func (t *joalTorrent) TrackerAnnounceUrls() []url.URL {
	uniqueRegistry := map[string]bool{}
	var urls []url.URL
	u, err := url.Parse(t.metaInfo.Announce)
	if err == nil {
		uniqueRegistry[u.String()] = true
		urls = append(urls, *u)
	}

	for a := range t.metaInfo.AnnounceList.DistinctValues() {
		if strings.TrimSpace(a) == "" {
			continue
		}
		u, err := url.Parse(a)
		if err != nil {
			continue
		}

		if _, contains := uniqueRegistry[u.String()]; contains {
			continue
		}
		uniqueRegistry[u.String()] = true
		urls = append(urls, *u)
	}

	return urls
}
func (t *joalTorrent) Size() int64 {
	return t.info.Length
}

func (t *joalTorrent) StartSeeding(client emulatedclient.IEmulatedClient, bandwidthClaimerPool bandwidth.IBandwidthClaimerPool) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.isRunning {
		return fmt.Errorf("already started")
	}
	t.isRunning = true

	currentSession := &seedSession{
		seedStats:   newSeedStats(),
		torrentName: t.info.Name,
		infoHash:    t.infoHash,
		swarm:       newSwarmElector(),
	}

	orhestra, err := client.CreateOrchestratorForTorrent(&orchestrator.TorrentInfo{
		Announce:     t.metaInfo.Announce,
		AnnounceList: t.metaInfo.AnnounceList.Clone(),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create orchestrator for torrent '%s'", t.path)
	}

	go func() {
		defer bandwidthClaimerPool.RemoveFromPool(currentSession)

		announceClosure := createAnnounceClosure(currentSession, client, bandwidthClaimerPool)
		orhestra.Start(announceClosure)

		stopRequest := <-t.stopping
		orhestra.Stop(stopRequest.ctx, announceClosure)
		stopRequest.doneStopping <- struct{}{}
	}()
	return nil
}

func (t *joalTorrent) StopSeeding(ctx context.Context) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.isRunning {
		return
	}
	t.isRunning = false

	log := logs.GetLogger().With(zap.String("torrent", filepath.Base(t.path)))

	stopRequest := &stoppingRequest{
		ctx:          ctx,
		doneStopping: make(chan struct{}),
	}
	log.Info("torrent: stopping")
	t.stopping <- stopRequest

	<-stopRequest.doneStopping
	log.Info("torrent: stopped")
}

func createAnnounceClosure(currentSession *seedSession, client emulatedclient.IEmulatedClient, bandwidthClaimerPool bandwidth.IBandwidthClaimerPool) orchestrator.AnnouncingFunction {
	log := logs.GetLogger()
	return func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {
		log.Info("announcing to tracker",
			zap.String("torrent", currentSession.torrentName),
			zap.String("event", event.String()),
			zap.String("uploaded", dataunit.ByteCountSI(currentSession.seedStats.Uploaded())),
			zap.String("tracker", u.Host),
			zap.ByteString("infohash", currentSession.infoHash[:]),
		)
		broadcast.EmitTorrentAnnouncing(broadcast.TorrentAnnouncingEvent{
			Infohash:      currentSession.infoHash,
			TrackerUrl:    u,
			AnnounceEvent: event,
			Uploaded:      currentSession.seedStats.Uploaded(),
		})
		resp, err := client.Announce(ctx, u, currentSession.InfoHash(), currentSession.seedStats.Uploaded(), currentSession.seedStats.Downloaded(), currentSession.seedStats.Left(), event)
		if err != nil {
			log.Warn("failed to announce", zap.String("tracker-url", u.String()), zap.Error(err))
			if event != tracker.Stopped {
				swarmHasChanged := currentSession.swarm.UpdateSwarm(errorSwarmUpdateRequest(u))
				if swarmHasChanged {
					broadcast.EmitTorrentSwarmChanged(broadcast.TorrentSwarmChangedEvent{
						Infohash: currentSession.infoHash,
						Seeder:   resp.Seeders,
						Leechers: resp.Leechers,
					})
					bandwidthClaimerPool.AddOrUpdate(currentSession)
				}
			}
			broadcast.EmitTorrentAnnounceFailed(broadcast.TorrentAnnounceFailedEvent{
				Infohash:      currentSession.infoHash,
				TrackerUrl:    u,
				AnnounceEvent: event,
				Datetime:      time.Now(),
				Error:         err.Error(),
			})
			return announcer.AnnounceResponse{}, errors.Wrap(err, "failed to announce")
		}
		broadcast.EmitTorrentAnnounceSuccess(broadcast.TorrentAnnounceSuccessEvent{
			Infohash:      currentSession.infoHash,
			TrackerUrl:    u,
			AnnounceEvent: event,
			Datetime:      time.Now(),
			Seeder:        resp.Seeders,
			Leechers:      resp.Leechers,
			Interval:      resp.Interval,
		})
		log.Info("tracker answered",
			zap.String("torrent", currentSession.torrentName),
			zap.String("tracker", u.Host),
			zap.String("interval", resp.Interval.String()),
			zap.Int32("leechers", resp.Leechers),
			zap.Int32("leechers", resp.Seeders),
		)

		if event != tracker.Stopped {
			swarmHasChanged := currentSession.swarm.UpdateSwarm(successSwarmUpdateRequest(u, resp))
			if swarmHasChanged {
				broadcast.EmitTorrentSwarmChanged(broadcast.TorrentSwarmChangedEvent{
					Infohash: currentSession.infoHash,
					Seeder:   resp.Seeders,
					Leechers: resp.Leechers,
				})
				bandwidthClaimerPool.AddOrUpdate(currentSession)
			}
		}

		return resp, nil
	}
}

type seedSession struct {
	seedStats   seedStats
	infoHash    torrent.InfoHash
	torrentName string
	swarm       *swarmElector
}

func (c *seedSession) InfoHash() torrent.InfoHash {
	return c.infoHash
}

func (c *seedSession) AddUploaded(bytes int64) {
	c.seedStats.AddUploaded(bytes)
}

func (c *seedSession) GetSwarm() bandwidth.ISwarm {
	return c.swarm
}

type seedStats interface {
	Uploaded() int64
	Downloaded() int64
	Left() int64
	AddUploaded(bytes int64)
}

type mutableSeedStats struct {
	uploaded   int64
	downloaded int64
	left       int64
}

func newSeedStats() *mutableSeedStats {
	return &mutableSeedStats{
		uploaded:   0,
		downloaded: 0,
		left:       0,
	}
}

func (m mutableSeedStats) Uploaded() int64 {
	return m.uploaded
}
func (m mutableSeedStats) Downloaded() int64 {
	return m.downloaded
}
func (m mutableSeedStats) Left() int64 {
	return m.left
}

func (m *mutableSeedStats) AddUploaded(bytes int64) {
	m.uploaded += bytes
}
