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
	AddUploaded(bytes int64)
	// May return nil
	GetSwarm() bandwidth.ISwarm
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

type seedingTorrent struct {
	seedStats
	path      string
	metaInfo  *slimMetaInfo
	info      *slimInfo
	infoHash  torrent.InfoHash
	swarm     bandwidth.ISwarm
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

	return &seedingTorrent{
		seedStats: seedStats{
			Downloaded: 0,
			Left:       0,
			Uploaded:   0,
		},
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
		swarm:     nil,
		isRunning: false,
		stopping:  make(chan chan struct{}),
		lock:      &sync.Mutex{},
	}, nil
}

func (t seedingTorrent) InfoHash() torrent.InfoHash {
	return t.infoHash
}

func (t seedingTorrent) GetSwarm() bandwidth.ISwarm {
	return t.swarm
}

func (t *seedingTorrent) StartSeeding(client emulatedclient.IEmulatedClient, dispatcher bandwidth.IDispatcher) {
	// TODO: create orchestrator & everything needed here and close with defer since this has to be a blocking function.
	panic("not implemented")
}

func (t *seedingTorrent) StopSeeding(ctx context.Context) {
	panic("not implemented")
	// TODO: Announce stop to all and wait for them to return before reseting to 0 (otherwise an announce may be sent with a 0 uploaded
	// TODO: reset swarm to 0
	// TODO: reset seed stats
}

func createAnnounceFunction(t *seedingTorrent, client emulatedclient.EmulatedClient, dispatcher bandwidth.IDispatcher) orchestrator.AnnouncingFunction {
	return func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {

		resp, err := client.Announce(u, t.infoHash, t.Uploaded, t.Downloaded, t.Left, event, ctx)
		if err != nil {
			return tracker.AnnounceResponse{}, errors.Wrap(err, "failed to announce")
		}
		//TODO: publish res & error (most likely create our own struct and publish to chan)

		// TODO: a tricky think to do will be to evaluate the real number of peer for a torrent since multiple tracker may return different peer count. url may be used to differentiate trackers response and maintain an average or max-peer count for each
		// TODO: this has to be called after the torrent has updated his seeders/leechers, but not to many times to prevent overhead calculations dispatcher.ClaimOrUpdate(t)

		return resp, nil
	}
}
