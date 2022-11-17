package usecases

import (
	"fmt"
	commonDomain "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"
	"github.com/anthonyraymond/joal-cli/internal/v3/sharing/domain"
)

type ShareTorrentUseCase interface {
	execute(id commonDomain.TorrentId) error
}

type ShareTorrentUseCaseImpl struct {
	repository SharedTorrentRepository
}

func (u ShareTorrentUseCaseImpl) execute(torrentId commonDomain.TorrentId) error {
	sharedTorrent := domain.SharedTorrent{
		TorrentId:   torrentId,
		Downloading: false,
		Seeding:     true,
	}

	err := u.repository.Save(sharedTorrent)
	if err != nil {
		return fmt.Errorf("failed to save SharedTorrent [%s]: %w", torrentId, err)
	}
	return nil
}
