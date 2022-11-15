package domain

import (
	"errors"
	v3 "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"
)

var (
	SharedTorrentNotFound = errors.New("shared torrent not found in repository")
)

type SharedTorrentRepository interface {
	FindByTorrentId(torrentId v3.TorrentId) (SharedTorrent, error)
	Remove(torrentId v3.TorrentId) error
	Save(SharedTorrent) error
}
