package sharing

import (
	"github.com/anacrolix/torrent/metainfo"
)

type TorrentId string

func (t TorrentId) FromInfoHash(hash metainfo.Hash) TorrentId {
	return TorrentId(hash.HexString())
}
