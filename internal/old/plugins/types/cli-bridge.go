package types

import "C"
import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/internal/old/core"
	"github.com/anthonyraymond/joal-cli/internal/old/core/broadcast"
	"github.com/anthonyraymond/joal-cli/internal/old/core/manager2"
	"io"
	"io/ioutil"
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
	manager      manager2.Manager
	configLoader *core.CoreConfigLoader
}

func NewCoreBridge(loader *core.CoreConfigLoader, manager manager2.Manager) ICoreBridge {
	return &coreBridge{
		manager:      manager,
		configLoader: loader,
	}
}

func (b *coreBridge) StartSeeding() error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}
	b.manager.StartSeeding()
	return nil
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
	broadcast.EmitConfigChanged(broadcast.ConfigChangedEvent{
		NeedRestartToTakeEffect: true,
		RuntimeConfig:           conf.RuntimeConfig,
	})

	return &RuntimeConfig{
		MinimumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds,
		MaximumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds,
		Client:                 conf.RuntimeConfig.Client,
	}, nil
}

func (b *coreBridge) AddTorrent(filename string, r io.Reader) error {
	// Extract the content from the http request reader since the manager.SaveTorrentFile is asynchronous
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read torrent file from reader: %w", err)
	}
	b.manager.SaveTorrentFile(filename, content)
	return nil
}

func (b *coreBridge) RemoveTorrent(infohash torrent.InfoHash) error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}

	b.manager.ArchiveTorrent(infohash)
	return nil
}

func (b *coreBridge) ListClientFiles() ([]string, error) {
	config, err := b.configLoader.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	files, err := ioutil.ReadDir(config.ClientsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list '%s' directory: %w", config.ClientsDir, err)
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
