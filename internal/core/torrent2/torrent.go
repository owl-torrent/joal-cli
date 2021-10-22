package torrent2

import (
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anthonyraymond/joal-cli/internal/core/announcer"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/utils/stop"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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
	trackers  []Tracker
	metaInfo  *slimMetaInfo
	info      *slimInfo
	isRunning bool
	stopping  stop.Chan
	lock      *sync.Mutex
}

func FromFile(filePath string, supportAnnounceList bool) (Torrent, error) {
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
			Private:     info.Private,
			Source:      info.Source,
		},
		trackers:  newTrackers(meta.Announce, meta.AnnounceList, supportAnnounceList),
		infoHash:  infoHash,
		isRunning: false,
		stopping:  stop.NewChan(),
		lock:      &sync.Mutex{},
	}, nil
}

func (t *torrentImpl) Start(client emulatedclient.IEmulatedClient) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.isRunning {
		return
	}
	t.isRunning = true

	// Disable trackers based on client capabilities (UDP, HTTP, ...)
	for _, track := range t.trackers {
		if strings.Contains(strings.ToLower(track.Url().Scheme), "http") && !client.SupportsHttpAnnounce() {
			track.Disable()
		} else if strings.Contains(strings.ToLower(track.Url().Scheme), "udp") && !client.SupportsUdpAnnounce() {
			track.Disable()
		}
	}

	go torrentRoutine(t, client)
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

func torrentRoutine(t *torrentImpl, client emulatedclient.IEmulatedClient) {
	onAnnounceSuccess := make(chan announcer.AnnounceResponse)
	onAnnounceError := make(chan announcer.AnnounceResponse)
	onAnnounceTime := time.After(0 * time.Second)

	for {
		select {
		case resp := <-onAnnounceSuccess:

		case errorResponse := <-onAnnounceError:

		case <-onAnnounceTime:

		case stopRequest := <-t.stopping:
			// TODO: announce stop with current stats

			t.peers.Reset()
			t.stats.Reset()
			stopRequest.NotifyDone()
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
	Private     *bool
	Source      string
}
