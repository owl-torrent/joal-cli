package seed

import (
	"context"
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
	StartSeeding(client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher)
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

type joalTorrent struct {
	ISeedSession
	path         string
	metaInfo     *slimMetaInfo
	info         *slimInfo
	infoHash     torrent.InfoHash
	isRunning    bool
	stopping     chan chan struct{}
	lock         *sync.Mutex
	orchestrator orchestrator.IOrchestrator
	swarm        *swarmElector
}

func FromReader(filePath string, client emulatedclient.IEmulatedClient) (ITorrent, error) {
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
	//TODO: move the tracker shuffling in orchestrator, it shouldn't be here
	for _, tier := range meta.AnnounceList {
		rand.Shuffle(len(tier), func(i, j int) {
			tier[i], tier[j] = tier[j], tier[i]
		})
	}

	o, err := client.CreateOrchestratorForTorrent(*meta)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create orchestrator for torrent '%s'", filePath)
	}

	return &joalTorrent{
		//TODO: init IseedSession
		swarm:        nil,
		orchestrator: o,
		path:         filePath,
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
		stopping:  make(chan chan struct{}),
		lock:      &sync.Mutex{},
	}, nil
}

func (t joalTorrent) InfoHash() torrent.InfoHash {
	return t.infoHash
}

func (t joalTorrent) GetSwarm() bandwidth.ISwarm {
	return t.swarm
}

func (t *joalTorrent) StartSeeding(client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) {
	t.lock.Lock()
	if t.isRunning {
		t.lock.Unlock()
		return
	}
	t.isRunning = true
	t.lock.Unlock()

	t.swarm = newSwarmElector()
	t.ISeedSession = newSeedSession()

	//announceClosure := createAnnounceClosure(t, client, dispatcher)

	// TODO: start orchestrator, swarm & everything needed here and create a goroutine

	// TODO: on stop, nil t.swarm & t.ISeedSession

	panic("not implemented")
}

func (t *joalTorrent) StopSeeding(ctx context.Context) {
	panic("not implemented")

	// TODO: send stop signal to main loop
}

func createAnnounceClosure(t *joalTorrent, client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) orchestrator.AnnouncingFunction {
	return func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {

		resp, err := client.Announce(ctx, u, t.InfoHash(), t.Uploaded(), t.Downloaded(), t.Left(), event)
		if err != nil {
			if event != tracker.Stopped {
				t.swarm.UpdateSwarm(errorSwarmUpdateRequest(u))
				if t.swarm.HasChanged() {
					t.swarm.ResetChanged()
					dispatcher.ClaimOrUpdate(t)
				}
			}
			return announcer.AnnounceResponse{}, errors.Wrap(err, "failed to announce")
		}

		if event != tracker.Stopped {
			t.swarm.UpdateSwarm(successSwarmUpdateRequest(u, resp))
			if t.swarm.HasChanged() {
				t.swarm.ResetChanged()
				dispatcher.ClaimOrUpdate(t)
			}
		}

		//TODO: publish res & error (most likely create our own struct and publish to chan)

		return resp, nil
	}
}

type ISeedSession interface {
	Uploaded() int64
	Downloaded() int64
	Left() int64
	AddUploaded(bytes int64)
}

type mutableSeedSession struct {
	uploaded   int64
	downloaded int64
	left       int64
}

func newSeedSession() *mutableSeedSession {
	return &mutableSeedSession{
		uploaded:   0,
		downloaded: 0,
		left:       0,
	}
}

func (m mutableSeedSession) Uploaded() int64 {
	return m.uploaded
}
func (m mutableSeedSession) Downloaded() int64 {
	return m.downloaded
}
func (m mutableSeedSession) Left() int64 {
	return m.left
}

func (m *mutableSeedSession) AddUploaded(bytes int64) {
	m.uploaded += bytes
}
