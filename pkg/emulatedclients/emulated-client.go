package emulatedclients

import (
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/announce"
	keygenerator "github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key/generator"
	peeridgenerator "github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid/generator"
)

type EmulatedClient struct {
	Name            string                           `yaml:"name"`
	Version         string                           `yaml:"version"`
	KeyGenerator    *keygenerator.KeyGenerator       `yaml:"keyGenerator"`
	PeerIdGenerator *peeridgenerator.PeerIdGenerator `yaml:"peerIdGenerator"`
	NumWant         int32                            `yaml:"numwant"`
	NumWantOnStop   int32                            `yaml:"numwantOnStop"`
	Announcer       *announce.Announcer              `yaml:"announcer"`
	Listener        *Listener                        `yaml:"listener"`
}

func (c *EmulatedClient) AfterPropertiesSet() error {
	if c.KeyGenerator == nil {
		return errors.New("'keyGenerator' key is required in EmulatedClient")
	}
	err := c.KeyGenerator.AfterPropertiesSet()
	if err != nil {
		return err
	}

	if c.PeerIdGenerator == nil {
		return errors.New("'peerIdGenerator' key is required in EmulatedClient")
	}
	err = c.PeerIdGenerator.AfterPropertiesSet()
	if err != nil {
		return err
	}

	if c.Announcer == nil {
		return errors.New("'announcer' key is required in EmulatedClient")
	}
	err = c.Announcer.AfterPropertiesSet()
	if err != nil {
		return err
	}

	if c.Listener == nil {
		return errors.New("'announcer' key is required in EmulatedClient")
	}
	err = c.Listener.AfterPropertiesSet()
	if err != nil {
		return err
	}
	return nil
}

func (c *EmulatedClient) Announce(infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
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

	return c.Announcer.Announce(nil, announceRequest)
}

func (c *EmulatedClient) StartListener() error {
	return c.Listener.Start()
}
