package types

import "C"
import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anthonyraymond/joal-cli/internal/core"
	"github.com/anthonyraymond/joal-cli/internal/core/seedmanager"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ICoreBridge interface {
	StartSeeding() error
	StopSeeding(ctx context.Context) error
	GetCoreConfig() (*RuntimeConfig, error)
	UpdateCoreConfig(config *RuntimeConfig) (*RuntimeConfig, error)
	AddTorrent(filename string, r io.Reader) error
	RemoveTorrent(infohash torrent.InfoHash) error
	ListClientFiles() ([]string, error)
}

type Config struct {
	NeedRestartToTakeEffect bool           `json:"needRestartToTakeEffect"`
	RuntimeConfig           *RuntimeConfig `json:"runtimeConfig"`
}

type RuntimeConfig struct {
	MinimumBytesPerSeconds int64  `json:"minimumBytesPerSeconds"`
	MaximumBytesPerSeconds int64  `json:"maximumBytesPerSeconds"`
	Client                 string `json:"client"`
}

type coreBridge struct {
	manager      seedmanager.ITorrentManager
	configLoader *core.CoreConfigLoader
}

func NewCoreBridge(loader *core.CoreConfigLoader) ICoreBridge {
	return &coreBridge{
		configLoader: loader,
	}
}

func (b *coreBridge) SetTorrentManager(manager seedmanager.ITorrentManager) {
	b.manager = manager
}

func (b *coreBridge) StartSeeding() error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}
	return b.manager.StartSeeding(nil)
}

func (b *coreBridge) StopSeeding(ctx context.Context) error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}
	b.manager.StopSeeding(ctx)

	return ctx.Err()
}

func (b *coreBridge) GetCoreConfig() (*RuntimeConfig, error) {
	if b.manager == nil {
		return nil, fmt.Errorf("torrent manager is not available yet")
	}
	conf, err := b.configLoader.ReadConfig()
	if err != nil {
		return nil, err
	}

	return &RuntimeConfig{
		MinimumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds,
		MaximumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds,
		Client:                 conf.RuntimeConfig.Client,
	}, nil
}

func (b *coreBridge) UpdateCoreConfig(newConf *RuntimeConfig) (*RuntimeConfig, error) {
	if b.manager == nil {
		return nil, fmt.Errorf("torrent manager is not available yet")
	}
	conf, err := b.configLoader.ReadConfig()
	if err != nil {
		return nil, err
	}
	conf.RuntimeConfig.Client = newConf.Client
	conf.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds = newConf.MinimumBytesPerSeconds
	conf.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds = newConf.MaximumBytesPerSeconds

	err = b.configLoader.SaveConfigToFile(conf.RuntimeConfig)
	if err != nil {
		return nil, err
	}

	return &RuntimeConfig{
		MinimumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds,
		MaximumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds,
		Client:                 conf.RuntimeConfig.Client,
	}, nil
}

func (b *coreBridge) AddTorrent(filename string, r io.Reader) error {
	meta, err := metainfo.Load(r)
	if err != nil {
		return errors.Wrap(err, "failed to parse torrent file")
	}

	config, err := b.configLoader.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read config file")
	}
	if filepath.Ext(filename) != ".torrent" {
		filename += ".torrent"
	}

	filename = filepath.Join(config.TorrentsDir, filename)
	w, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to open file '%s' for writing", filename)
	}

	err = meta.Write(w)
	if err != nil {
		return errors.Wrapf(err, "failed to write to file '%s'", filename)
	}

	return nil
}

func (b *coreBridge) RemoveTorrent(infohash torrent.InfoHash) error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}

	return b.manager.RemoveTorrent(infohash)
}

func (b *coreBridge) ListClientFiles() ([]string, error) {
	config, err := b.configLoader.ReadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	files, err := ioutil.ReadDir(config.ClientsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list '%s' directory", config.ClientsDir)
	}

	var clients []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != "yml" && ext != "yaml" {
			continue
		}
		clients = append(clients, file.Name())
	}
	return clients, nil
}
