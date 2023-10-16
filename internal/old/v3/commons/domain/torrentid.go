package domain

import "github.com/anacrolix/torrent"

type TorrentId struct {
	infoHash         torrent.InfoHash
	infoHashAsHexStr string // Store the hex representation cached in memory to avoid re-calculation on every comparison
}

func NewTorrentId(hash torrent.InfoHash) TorrentId {
	return TorrentId{
		infoHash:         hash,
		infoHashAsHexStr: hash.HexString(),
	}
}

func (t TorrentId) InfoHash() torrent.InfoHash {
	return t.infoHash
}

func (t TorrentId) Equals(t2 TorrentId) bool {
	return t.infoHashAsHexStr == t2.infoHashAsHexStr
}

func (t TorrentId) String() string {
	return t.infoHashAsHexStr
}
