package web

type ICoreBridge interface {
	StartSeeding() error
	StopSeeding() error
	UpdateConfig(config *Config) (Config, error)
	RemoveTorrent(infohash string) error
	AddTorrent(file []byte) error
}
