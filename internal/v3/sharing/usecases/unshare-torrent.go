package usecases

import (
	"errors"
	"fmt"
	commonDomain "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"
)

type UnShareTorrentUseCase interface {
	execute(id commonDomain.TorrentId) error
}

type UnShareTorrentUseCaseImpl struct {
	repository SharedTorrentRepository
}

func (u UnShareTorrentUseCaseImpl) execute(torrentId commonDomain.TorrentId) error {
	err := u.repository.Remove(torrentId)
	if err != nil {
		if errors.Is(err, SharedTorrentNotFound) {
			return nil
		}
		return fmt.Errorf("failed to remove SharedTorrent [%s] from repository: %w", torrentId, err)
	}
	return nil
}
