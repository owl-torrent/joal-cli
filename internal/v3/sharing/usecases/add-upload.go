package usecases

import (
	"errors"
	"fmt"
	v3 "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"
)

type AddUploadUseCase interface {
	execute(torrentId v3.TorrentId, bytesDownloaded int64) error
}

type AddUploadUseCaseImpl struct {
	repository SharedTorrentRepository
}

func (u AddUploadUseCaseImpl) execute(torrentId v3.TorrentId, bytesDownloaded int64) error {
	sharedTorrent, err := u.repository.FindByTorrentId(torrentId)
	if err != nil {
		if errors.Is(err, SharedTorrentNotFound) {
			return nil
		}
		return fmt.Errorf("failed to query SharedTorrent [%s] from repository: %w", torrentId, err)
	}
	err = sharedTorrent.AddUploaded(bytesDownloaded)
	if err != nil {
		return fmt.Errorf("failed to add downloaded to torrent [%s]: %w", torrentId, err)
	}

	err = u.repository.Save(sharedTorrent)
	if err != nil {
		return fmt.Errorf("failed to save SharedTorrent [%s]: %w", torrentId, err)
	}
	return nil
}
