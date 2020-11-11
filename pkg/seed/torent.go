package seed

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/announcer"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
	"github.com/anthonyraymond/joal-cli/pkg/orchestrator"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/rand"
	"net/url"
	"path/filepath"
	"sync"
)

type ITorrent interface {
	InfoHash() torrent.InfoHash
	StartSeeding(client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) error
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

func FromReader(filePath string) (ITorrent, error) {
	log := logs.GetLogger()
	meta, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load meta-info from file '%s'", filePath)
	}

	info, err := meta.UnmarshalInfo()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load info from file '%s'", filePath)
	}
	infoHash := meta.HashInfoBytes()
	log.Info("torrent parsed", zap.String("torrent", filepath.Base(filePath)), zap.ByteString("infohash", infoHash.Bytes()))

	// TODO: move the tracker shuffling in orchestrator, it shouldn't be here
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

func (t joalTorrent) InfoHash() torrent.InfoHash {
	return t.infoHash
}

func (t *joalTorrent) StartSeeding(client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.isRunning {
		return fmt.Errorf("already started")
	}
	t.isRunning = true

	currentSession := &seedSession{
		seedStats: newSeedStats(),
		infoHash:  t.infoHash,
		swarm:     newSwarmElector(),
	}

	orhestra, err := client.CreateOrchestratorForTorrent(&orchestrator.TorrentInfo{
		Announce:     t.metaInfo.Announce,
		AnnounceList: t.metaInfo.AnnounceList.Clone(),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create orchestrator for torrent '%s'", t.path)
	}

	go func() {
		defer dispatcher.Release(currentSession)

		announceClosure := createAnnounceClosure(currentSession, client, dispatcher)
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

	stopRequest := &stoppingRequest{
		ctx:          ctx,
		doneStopping: make(chan struct{}),
	}
	t.stopping <- stopRequest

	<-stopRequest.doneStopping
}

func createAnnounceClosure(currentSession *seedSession, client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) orchestrator.AnnouncingFunction {
	return func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {

		resp, err := client.Announce(ctx, u, currentSession.InfoHash(), currentSession.seedStats.Uploaded(), currentSession.seedStats.Downloaded(), currentSession.seedStats.Left(), event)
		if err != nil {
			if event != tracker.Stopped {
				currentSession.swarm.UpdateSwarm(errorSwarmUpdateRequest(u))
				if currentSession.swarm.HasChanged() {
					currentSession.swarm.ResetChanged()
					dispatcher.ClaimOrUpdate(currentSession)
				}
			}
			return announcer.AnnounceResponse{}, errors.Wrap(err, "failed to announce")
		}

		if event != tracker.Stopped {
			currentSession.swarm.UpdateSwarm(successSwarmUpdateRequest(u, resp))
			if currentSession.swarm.HasChanged() {
				currentSession.swarm.ResetChanged()
				dispatcher.ClaimOrUpdate(currentSession)
			}
		}

		return resp, nil
	}
}

type seedSession struct {
	seedStats seedStats
	infoHash  torrent.InfoHash
	swarm     *swarmElector
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
