package types

import "C"
import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/core"
	"github.com/anthonyraymond/joal-cli/internal/core/seedmanager"
)

type ICoreBridge interface {
	StartSeeding() error
	StopSeeding(ctx context.Context) error
	GetCoreConfig() (*Config, error)
	UpdateCoreConfig(config *RuntimeConfig) (*Config, error)
	RemoveTorrent(infohash string) error
	AddTorrent(file []byte) error
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
	return b.manager.StartSeeding()
}

func (b *coreBridge) StopSeeding(ctx context.Context) error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}
	b.manager.StopSeeding(ctx)

	return ctx.Err()
}

func (b *coreBridge) GetCoreConfig() (*Config, error) {
	if b.manager == nil {
		return nil, fmt.Errorf("torrent manager is not available yet")
	}
	conf, err := b.configLoader.ReadConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		NeedRestartToTakeEffect: true, // TODO: this should come back from the configloader and not being a fixed value
		RuntimeConfig: &RuntimeConfig{
			MinimumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds,
			MaximumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds,
			Client:                 conf.RuntimeConfig.Client,
		},
	}, nil
}

func (b *coreBridge) UpdateCoreConfig(newConf *RuntimeConfig) (*Config, error) {
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

	return &Config{
		NeedRestartToTakeEffect: true, // TODO: this should come back from the configloader and not being a fixed value
		RuntimeConfig: &RuntimeConfig{
			MinimumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds,
			MaximumBytesPerSeconds: conf.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds,
			Client:                 conf.RuntimeConfig.Client,
		},
	}, nil
}

func (b *coreBridge) AddTorrent(file []byte) error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}
	// TODO: implement
	panic("implement me")
}

func (b *coreBridge) RemoveTorrent(infohash string) error {
	if b.manager == nil {
		return fmt.Errorf("torrent manager is not available yet")
	}
	// TODO: implement
	panic("implement me")
}
