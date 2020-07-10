package seed

import (
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient"
	"github.com/anthonyraymond/joal-cli/pkg/orchestrator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	path      string
	metaInfo  *slimMetaInfo
	info      *slimInfo
	infoHash  torrent.InfoHash
	isRunning bool
	stopping  chan chan struct{}
	lock      *sync.Mutex
}

func FromReader(filePath string) (ITorrent, error) {
	meta, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load meta-info from file '%s'", filePath)
	}

	info, err := meta.UnmarshalInfo()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load info from file '%s'", filePath)
	}
	infoHash := meta.HashInfoBytes()
	logrus.WithFields(logrus.Fields{
		"torrent":  filepath.Base(filePath),
		"infohash": infoHash,
	}).Info("torrent parsed")

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
		stopping:  make(chan chan struct{}),
		lock:      &sync.Mutex{},
	}, nil
}

func (t joalTorrent) InfoHash() torrent.InfoHash {
	return t.infoHash
}

func (t *joalTorrent) StartSeeding(client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) {
	// TODO: create orchestrator, swarm & everything needed here and close with defer since this has to be a blocking function.
	//  doing this will prevent weird state and concurrent access to struct attributes.
	//  Swarm, orchestrator and so on will only exists in the scope of this function, no need to reset swarm, and so on.
	panic("not implemented")
}

func (t *joalTorrent) StopSeeding(ctx context.Context) {
	panic("not implemented")
	// TODO: Announce stop to all and wait for them to return before reseting to 0 (otherwise an announce may be sent with a 0 uploaded
	// TODO: reset seed stats
}

func createAnnounceClosure(client emulatedclient.EmulatedClient, dispatcher bandwidth.IDispatcher, seedSession *mutableSeedSession) orchestrator.AnnouncingFunction {
	return func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {

		resp, err := client.Announce(u, seedSession.InfoHash(), seedSession.Uploaded(), seedSession.Downloaded(), seedSession.Left(), event, ctx)
		if err != nil {
			if event != tracker.Stopped {
				seedSession.swarm.UpdateSwarm(errorSwarmUpdateRequest(u))
				if seedSession.swarm.HasChanged() {
					seedSession.swarm.ResetChanged()
					dispatcher.ClaimOrUpdate(seedSession)
				}
			}
			return tracker.AnnounceResponse{}, errors.Wrap(err, "failed to announce")
		}

		if event != tracker.Stopped {
			seedSession.swarm.UpdateSwarm(successSwarmUpdateRequest(u, resp))
			if seedSession.swarm.HasChanged() {
				seedSession.swarm.ResetChanged()
				dispatcher.ClaimOrUpdate(seedSession)
			}
		}

		//TODO: publish res & error (most likely create our own struct and publish to chan)

		return resp, nil
	}
}

type ISeedSession interface {
	InfoHash() torrent.InfoHash
	GetSwarm() bandwidth.ISwarm
	Uploaded() int32
	Downloaded() int32
	Left() int32
}

type mutableSeedSession struct {
	infoHash   torrent.InfoHash
	swarm      *swarmElector
	uploaded   int64
	downloaded int64
	left       int64
}

func (m mutableSeedSession) InfoHash() torrent.InfoHash {
	return m.infoHash
}
func (m mutableSeedSession) GetSwarm() bandwidth.ISwarm {
	return m.swarm.CurrentSwarm()
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
