package emulatedclients

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/announce"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/generator"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/urlencoder"
)

type EmulatedClient struct {
	Name            string                `yaml:"name"`
	Version         string                `yaml:"version"`
	KeyGenerator    generator.Generator   `yaml:"keyGenerator"`
	PeerIdGenerator generator.Generator   `yaml:"peerIdGenerator"`
	UrlEncoder      urlencoder.UrlEncoder `yaml:"urlEncoder"`
	Announcer       announce.Announcer    `yaml:"announcer"`
}
