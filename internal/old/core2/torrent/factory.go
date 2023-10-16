package torrent

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"net/url"
)

type FactoryImpl struct {
}

func (f FactoryImpl) CreateTorrent(meta metainfo.MetaInfo, announcePolicy AnnouncePolicy) (Torrent, error) {
	info, err := meta.UnmarshalInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal torrent info: %w", err)
	}

	announce, err := url.Parse(meta.Announce)
	if err != nil {
		return nil, fmt.Errorf("failed to parse announce url '%s': %w", meta.Announce, err)
	}
	announceList, err := parseAnnounceList(meta.AnnounceList)
	if err != nil {
		return nil, fmt.Errorf("failed to parse announce list: %w", err)
	}

	trackers, err := createTrackers(*announce, announceList, announcePolicy)

	return &torrent{
		infoHash: meta.HashInfoBytes(),
		info:     infoToSlimInfo(info),
		contrib:  contribution{},
		peers:    newPeersElector(mostLeeched),
		trackers: trackers,
	}, nil
}

func parseAnnounceList(announceList [][]string) ([][]url.URL, error) {
	result := make([][]url.URL, len(announceList))

	for tier := range announceList {
		result[tier] = make([]url.URL, len(announceList[tier]))

		for track := range announceList[tier] {
			u, err := url.Parse(announceList[tier][track])
			if err != nil {
				return nil, fmt.Errorf("failed to parse url in announceList '%s': %w", announceList[tier][track], err)
			}
			result[tier][track] = *u
		}
	}

	return result, nil
}
