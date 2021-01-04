package plugins

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/core/seedmanager"
)

type ICoreBridge interface {
	StartSeeding() error
	StopSeeding(ctx context.Context) error
	UpdateCoreConfig(config *RuntimeConfig) (Config, error)
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
	manager seedmanager.ITorrentManager
}

func NewCoreBridge(manager seedmanager.ITorrentManager) ICoreBridge {
	return &coreBridge{
		manager: manager,
	}
}

func (b *coreBridge) StartSeeding() error {
	return b.manager.StartSeeding()
}

func (b *coreBridge) StopSeeding(ctx context.Context) error {
	b.manager.StopSeeding(ctx)

	return ctx.Err()
}

func (b *coreBridge) UpdateCoreConfig(config *RuntimeConfig) (Config, error) {
	// TODO: implement
	panic("implement me")
}

func (b *coreBridge) RemoveTorrent(infohash string) error {
	// TODO: implement
	panic("implement me")
}

func (b *coreBridge) AddTorrent(file []byte) error {
	// TODO: implement
	panic("implement me")
}
