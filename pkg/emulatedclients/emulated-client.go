package emulatedclients

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/announce"
	keygenerator "github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key/generator"
	peeridgenerator "github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid/generator"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/urlencoder"
)

type EmulatedClient struct {
	Name            string                          `yaml:"name"`
	Version         string                          `yaml:"version"`
	KeyGenerator    keygenerator.KeyGenerator       `yaml:"keyGenerator"`
	PeerIdGenerator peeridgenerator.PeerIdGenerator `yaml:"peerIdGenerator"`
	UrlEncoder      urlencoder.UrlEncoder           `yaml:"urlEncoder"`
	Announcer       announce.Announcer              `yaml:"announcer"`
}
