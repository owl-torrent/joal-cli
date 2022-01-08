package emulatedclient

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/announcer"
	"github.com/anthonyraymond/joal-cli/internal/core/announces"
	keygenerator "github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key/generator"
	peeridgenerator "github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/peerid/generator"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
)

type IEmulatedClient interface {
	GetName() string
	GetVersion() string
	Announce(request *announces.AnnounceRequest)
	StartListener(proxyFunc func(*http.Request) (*url.URL, error)) error
	StopListener(ctx context.Context)
	GetAnnounceCapabilities() AnnounceCapabilities
	SupportsHttpAnnounce() bool
	SupportsUdpAnnounce() bool
}

type EmulatedClient struct {
	Name                 string                           `yaml:"name" validate:"required"`
	Version              string                           `yaml:"version" validate:"required"`
	KeyGenerator         *keygenerator.KeyGenerator       `yaml:"keyGenerator" validate:"required"`
	PeerIdGenerator      *peeridgenerator.PeerIdGenerator `yaml:"peerIdGenerator" validate:"required"`
	NumWant              int32                            `yaml:"numwant" validate:"min=1"`
	NumWantOnStop        int32                            `yaml:"numwantOnStop"`
	AnnounceCapabilities AnnounceCapabilities             `yaml:"announceCapabilities" validate:"required"`
	Announcer            *announcer.Announcer             `yaml:"announcer" validate:"required"`
	Listener             *Listener                        `yaml:"listener" validate:"required"`
}

func FromClientFile(path string, proxyFunc func(*http.Request) (*url.URL, error)) (IEmulatedClient, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return FromReader(file, proxyFunc)
}

func FromReader(reader io.Reader, proxyFunc func(*http.Request) (*url.URL, error)) (IEmulatedClient, error) {
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

	err = client.AfterPropertiesSet(proxyFunc)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *EmulatedClient) AfterPropertiesSet(proxyFunc func(*http.Request) (*url.URL, error)) error {
	err := c.KeyGenerator.AfterPropertiesSet()
	if err != nil {
		return err
	}

	err = c.PeerIdGenerator.AfterPropertiesSet()
	if err != nil {
		return err
	}

	err = c.Announcer.AfterPropertiesSet(proxyFunc)
	if err != nil {
		return err
	}

	err = c.Listener.AfterPropertiesSet()
	if err != nil {
		return err
	}
	return nil
}

func (c *EmulatedClient) GetName() string {
	return c.Name
}

func (c *EmulatedClient) GetVersion() string {
	return c.Version
}

func (c *EmulatedClient) Announce(request *announces.AnnounceRequest) {
	if c.Listener.ip == nil || c.Listener.listeningPort == nil {
		panic(fmt.Errorf("EmulatedClient listener is not started"))
	}

	announceRequest := announcer.AnnounceRequest{
		InfoHash:   request.InfoHash,
		PeerId:     c.PeerIdGenerator.Get(request.InfoHash, request.Event),
		Downloaded: request.Downloaded,
		Left:       request.Left,
		Uploaded:   request.Uploaded,
		Corrupt:    request.Corrupt,
		Event:      request.Event,
		IPAddress:  *c.Listener.ip,
		Key:        uint32(c.KeyGenerator.Get(request.InfoHash, request.Event)),
		NumWant:    c.NumWant,
		Private:    request.Private,
		Port:       *c.Listener.listeningPort,
	}
	if request.Event == tracker.Stopped {
		announceRequest.NumWant = c.NumWantOnStop
	}

	response, err := c.Announcer.Announce(request.Url, announceRequest, request.Ctx)
	if err != nil {
		request.AnnounceCallbacks.Failed(announces.AnnounceResponseError{
			Request:  request,
			Error:    fmt.Errorf("announce failed: %w", err),
			Interval: 0,
		})
		return
	}
	request.AnnounceCallbacks.Success(announces.AnnounceResponse{
		Request:  request,
		Interval: response.Interval,
		Leechers: response.Leechers,
		Seeders:  response.Seeders,
		Peers:    response.Peers,
	})
}

func (c *EmulatedClient) StartListener(proxyFunc func(*http.Request) (*url.URL, error)) error {
	return c.Listener.Start(proxyFunc)
}

func (c *EmulatedClient) StopListener(ctx context.Context) {
	c.Listener.Stop(ctx)
}

func (c *EmulatedClient) GetAnnounceCapabilities() AnnounceCapabilities {
	return c.AnnounceCapabilities
}

func (c *EmulatedClient) SupportsHttpAnnounce() bool {
	return c.Announcer.Http != nil
}

func (c *EmulatedClient) SupportsUdpAnnounce() bool {
	return c.Announcer.Udp != nil
}
