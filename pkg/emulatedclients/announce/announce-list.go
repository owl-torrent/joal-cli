package announce

import (
	"github.com/anacrolix/torrent/metainfo"
)

func promoteTier(al *metainfo.AnnounceList, tier int) {
	if tier == 0 {
		return
	}
	newAnnounceList := make(metainfo.AnnounceList, len(*al))
	newAnnounceList[0] = (*al)[tier]
	for i, a := range *al {
		if i == tier {
			continue
		}
		if i < tier {
			newAnnounceList[i+1] = a
		} else {
			newAnnounceList[i] = a
		}
	}

	*al = newAnnounceList
}

func promoteUrlInTier(al *metainfo.AnnounceList, tier int, urlIndex int) {
	if urlIndex == 0 {
		return
	}
	currentTier := &(*al)[tier]
	newAnnounceUrls := make([]string, len(*currentTier))
	newAnnounceUrls[0] = (*currentTier)[urlIndex]
	for i, u := range *currentTier {
		if i == urlIndex {
			continue
		}
		if i < urlIndex {
			newAnnounceUrls[i+1] = u
		} else {
			newAnnounceUrls[i] = u
		}
	}

	*currentTier = newAnnounceUrls
}
