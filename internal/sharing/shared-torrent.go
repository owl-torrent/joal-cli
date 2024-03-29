package sharing

type State int

const (
	Downloading State = 0
	Seeding     State = 1
	Paused      State = 2
)

type SharedTorrent interface {
}

type sharedTorrentImpl struct {
	torrentId     TorrentId
	swarm         Swarm
	contributions Contributions
	state         State
}
