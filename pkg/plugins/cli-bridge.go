package plugins

import "context"

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
