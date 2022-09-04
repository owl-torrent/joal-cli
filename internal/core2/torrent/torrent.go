package torrent

type Torrent struct {
	contrib Contribution
	peers   peersElector
}
