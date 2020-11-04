package emulatedclient

//go:generate mockgen -destination=./emulated-client_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/emulatedclient -package=emulatedclient github.com/anthonyraymond/joal-cli/pkg/emulatedclient IEmulatedClient

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient/announce"
	keygenerator "github.com/anthonyraymond/joal-cli/pkg/emulatedclient/key/generator"
	peeridgenerator "github.com/anthonyraymond/joal-cli/pkg/emulatedclient/peerid/generator"
	"github.com/anthonyraymond/joal-cli/pkg/orchestrator"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
)

type IEmulatedClient interface {
	Announce(ctx context.Context, u url.URL, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error)
	StartListener() error
	StopListener(ctx context.Context)
	CreateOrchestratorForTorrent(info metainfo.MetaInfo) (orchestrator.IOrchestrator, error)
}

type EmulatedClient struct {
	Name                string                           `yaml:"name" validate:"required"`
	Version             string                           `yaml:"version" validate:"required"`
	KeyGenerator        *keygenerator.KeyGenerator       `yaml:"keyGenerator" validate:"required"`
	PeerIdGenerator     *peeridgenerator.PeerIdGenerator `yaml:"peerIdGenerator" validate:"required"`
	NumWant             int32                            `yaml:"numwant" validate:"min=1"`
	NumWantOnStop       int32                            `yaml:"numwantOnStop"`
	OrchestratorFactory *orchestratorFactory             `yaml:"announceOrchestrator" validate:"required"`
	Announcer           *announce.Announcer              `yaml:"announcer" validate:"required"`
	Listener            *Listener                        `yaml:"listener" validate:"required"`
}

func FromClientFile(path string) (IEmulatedClient, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return FromReader(file)
}

func FromReader(reader io.Reader) (IEmulatedClient, error) {
	client := EmulatedClient{}
	err := yaml.NewDecoder(reader).Decode(&client)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	validate.RegisterTagNameFunc(TagNameFunction) // add json capability
	err = validate.Struct(client)
	if err != nil {
		return nil, err
	}

	err = client.AfterPropertiesSet()
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *EmulatedClient) AfterPropertiesSet() error {
	err := c.KeyGenerator.AfterPropertiesSet()
	if err != nil {
		return err
	}

	err = c.PeerIdGenerator.AfterPropertiesSet()
	if err != nil {
		return err
	}

	err = c.Announcer.AfterPropertiesSet()
	if err != nil {
		return err
	}

	err = c.Listener.AfterPropertiesSet()
	if err != nil {
		return err
	}
	return nil
}

func (c *EmulatedClient) Announce(ctx context.Context, u url.URL, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
	if c.Listener.ip == nil || c.Listener.listeningPort == nil {
		panic(errors.New("EmulatedClient listener is not started"))
	}
	announceRequest := announce.AnnounceRequest{
		InfoHash:   infoHash,
		PeerId:     c.PeerIdGenerator.Get(infoHash, event),
		Downloaded: downloaded,
		Left:       left,
		Uploaded:   uploaded,
		Event:      event,
		IPAddress:  *c.Listener.ip,
		Key:        uint32(c.KeyGenerator.Get(infoHash, event)),
		NumWant:    c.NumWant,
		Port:       *c.Listener.listeningPort,
	}
	if event == tracker.Stopped {
		announceRequest.NumWant = c.NumWantOnStop
	}

	return c.Announcer.Announce(u, announceRequest, ctx)
}

func (c *EmulatedClient) StartListener() error {
	return c.Listener.Start()
}

func (c *EmulatedClient) StopListener(ctx context.Context) {
	c.Listener.Stop(ctx)
}

func (c *EmulatedClient) CreateOrchestratorForTorrent(info metainfo.MetaInfo) (orchestrator.IOrchestrator, error) {
	return c.OrchestratorFactory.createOrchestrator(info)
}
