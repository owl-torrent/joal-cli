package web

type CliBridge interface {
	StartSeeding() error
	StopSeeding() error
	UpdateConfig(config *Config) (Config, error)
	RemoveTorrent(infohash string) error
	AddTorrent(file []byte) error
}
