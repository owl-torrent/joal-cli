package usecases

import (
	"errors"
	"fmt"
	v3 "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"
	commonEvents "github.com/anthonyraymond/joal-cli/internal/v3/commons/events"
	"github.com/anthonyraymond/joal-cli/internal/v3/sharing/domain"
	"github.com/anthonyraymond/joal-cli/internal/v3/sharing/events"
)

type AddDownloadUseCase interface {
	execute(torrentId v3.TorrentId, bytesDownloaded int64) error
}

type AddDownloadUseCaseImpl struct {
	repository     domain.SharedTorrentRepository
	eventPublisher commonEvents.EventPublisher
}

func (u AddDownloadUseCaseImpl) execute(torrentId v3.TorrentId, bytesDownloaded int64) error {
	sharedTorrent, err := u.repository.FindByTorrentId(torrentId)
	if err != nil {
		if errors.Is(err, domain.SharedTorrentNotFound) {
			return nil
		}
		return fmt.Errorf("failed to query SharedTorrent [%s] from repository: %w", torrentId, err)
	}

	if !sharedTorrent.Downloading {
		return nil
	}

	err = sharedTorrent.AddUploaded(bytesDownloaded)
	if err != nil {
		return fmt.Errorf("failed to add uploaded to torrent [%s]: %w", torrentId, err)
	}

	if sharedTorrent.Downloading && sharedTorrent.IsDownloadComplete() {
		sharedTorrent.Downloading = false
		defer u.eventPublisher.Publish(events.NewTorrentDownloaded(sharedTorrent.TorrentId))
	}

	err = u.repository.Save(sharedTorrent)
	if err != nil {
		return fmt.Errorf("failed to save SharedTorrent [%s]: %w", torrentId, err)
	}

	return nil
}
