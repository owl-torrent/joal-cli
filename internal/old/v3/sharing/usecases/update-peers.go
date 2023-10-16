package usecases

import (
	"errors"
	"fmt"
	commonDomain "github.com/anthonyraymond/joal-cli/internal/old/v3/commons/domain"
)

type UpdatePeersUseCase interface {
	execute(torrentId commonDomain.TorrentId, peers commonDomain.Peers) error
}

type UpdatePeersUseCaseImpl struct {
	repository            SharedTorrentRepository
	peerElectionAlgorithm PeerElectionAlgorithm
}

type PeerElectionAlgorithm = func(p1 commonDomain.Peers, p2 commonDomain.Peers) commonDomain.Peers

var (
	mostLeeched PeerElectionAlgorithm = func(candidate1 commonDomain.Peers, candidate2 commonDomain.Peers) commonDomain.Peers {
		// The two below conditions ensure that the returned peer won't have 0 seeders if one of the two candidate has 1 seeders
		if candidate1.Seeders == 0 && candidate2.Seeders > 0 {
			return candidate2
		} else if candidate2.Seeders == 0 && candidate1.Seeders > 0 {
			return candidate1
		}
		if candidate1.Leechers < candidate2.Leechers {
			return candidate2
		}
		return candidate1
	}
)

func (u UpdatePeersUseCaseImpl) execute(torrentId commonDomain.TorrentId, peers []commonDomain.Peers) error {
	sharedTorrent, err := u.repository.FindByTorrentId(torrentId)
	if err != nil {
		if errors.Is(err, SharedTorrentNotFound) {
			return nil
		}
		return fmt.Errorf("failed to query SharedTorrent [%s] from repository: %w", torrentId, err)
	}
	if len(peers) == 0 {
		sharedTorrent.SetPeers(commonDomain.Peers{})
	}
	sharedTorrent.SetPeers(proceedToPeerElection(peers, u.peerElectionAlgorithm))

	err = u.repository.Save(sharedTorrent)
	if err != nil {
		return fmt.Errorf("failed to save SharedTorrent [%s]: %w", torrentId, err)
	}
	return nil
}

func proceedToPeerElection(candidates []commonDomain.Peers, elect PeerElectionAlgorithm) commonDomain.Peers {
	var elected commonDomain.Peers
	if len(candidates) == 0 {
		return elected
	}

	for _, candidate := range candidates {
		// if currently elected is zero value assign and continue
		if elected.Leechers == 0 && elected.Seeders == 0 {
			elected = candidate
			continue
		}
		elected = elect(elected, candidate)
	}
	return elected
}
