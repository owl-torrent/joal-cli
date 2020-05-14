package tmp

import (
	"github.com/anacrolix/torrent/metainfo"
)

type Torrent struct {
	metainfo metainfo.MetaInfo
	info metainfo.Info
}