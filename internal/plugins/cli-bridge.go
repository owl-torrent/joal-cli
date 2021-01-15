package plugins

import (
	"context"
	"github.com/anthonyraymond/joal-cli/internal/core/config"
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
	configLoader config.IConfigLoader
}

func NewCoreBridge(manager seedmanager.ITorrentManager, loader config.IConfigLoader) ICoreBridge {
	return &coreBridge{
		manager:      manager,
		configLoader: loader,
	}
}

func (b *coreBridge) StartSeeding() error {
	return b.manager.StartSeeding()
}

func (b *coreBridge) StopSeeding(ctx context.Context) error {
	b.manager.StopSeeding(ctx)

	return ctx.Err()
}

func (b *coreBridge) GetCoreConfig() (*Config, error) {
	conf, err := b.configLoader.LoadConfigAndInitIfNeeded()
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
	conf, err := b.configLoader.LoadConfigAndInitIfNeeded()
	if err != nil {
		return nil, err
	}
	conf.RuntimeConfig.Client = newConf.Client
	conf.RuntimeConfig.BandwidthConfig.Speed.MinimumBytesPerSeconds = newConf.MinimumBytesPerSeconds
	conf.RuntimeConfig.BandwidthConfig.Speed.MaximumBytesPerSeconds = newConf.MaximumBytesPerSeconds

	savedConf, err := b.configLoader.UpdateConfig(conf.RuntimeConfig)
	if err != nil {
		return nil, err
	}

	return &Config{
		NeedRestartToTakeEffect: true, // TODO: this should come back from the configloader and not being a fixed value
		RuntimeConfig: &RuntimeConfig{
			MinimumBytesPerSeconds: savedConf.BandwidthConfig.Speed.MinimumBytesPerSeconds,
			MaximumBytesPerSeconds: savedConf.BandwidthConfig.Speed.MaximumBytesPerSeconds,
			Client:                 savedConf.Client,
		},
	}, nil
}

func (b *coreBridge) AddTorrent(file []byte) error {
	// TODO: implement
	panic("implement me")
}

func (b *coreBridge) RemoveTorrent(infohash string) error {
	// TODO: implement
	panic("implement me")
}
