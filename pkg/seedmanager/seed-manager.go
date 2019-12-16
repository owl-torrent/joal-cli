package seedmanager

import (
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients"
	"github.com/anthonyraymond/joal-cli/pkg/seed"
	"github.com/anthonyraymond/joal-cli/pkg/seedmanager/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/anthonyraymond/watcher"
)

type SeedManager struct {
	conf                *config.SeedConfig
	joalPaths           *joalPaths
	seeds               map[torrent.InfoHash]seed.ISeed
	torrentFileWatcher  *watcher.Watcher
	bandwidthDispatcher bandwidth.IDispatcher
	client              emulatedclients.IEmulatedClient
	fileWatcherPoll     time.Duration
	lock                *sync.Mutex
}

func SeedManagerNew(joalPaths *joalPaths, conf config.SeedConfig) (*SeedManager, error) {
	dispatcher := bandwidth.DispatcherNew(&bandwidth.RandomSpeedProvider{
		MinimumBytesPerSeconds: conf.MinUploadRate,
		MaximumBytesPerSeconds: conf.MaxUploadRate,
	})

	client, err := emulatedclients.FromClientFile(path.Join(joalPaths.clientFileFolder, conf.Client))
	if err != nil {
		return nil, err
	}

	return &SeedManager{
		conf:                &conf,
		joalPaths:           joalPaths,
		seeds:               make(map[torrent.InfoHash]seed.ISeed),
		torrentFileWatcher:  nil,
		bandwidthDispatcher: dispatcher,
		client:              client,
		fileWatcherPoll:     1 * time.Second,
		lock:                &sync.Mutex{},
	}, nil
}

func (s *SeedManager) Start() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	torrentFileWatcher := watcher.New()
	torrentFileWatcher.AddFilterHook(torrentFileFilter())
	err := torrentFileWatcher.Add(s.joalPaths.torrentFolder)
	if err != nil {
		return err
	}
	s.torrentFileWatcher = torrentFileWatcher

	err = s.client.StartListener()
	if err != nil {
		return err
	}
	s.bandwidthDispatcher.Start()

	// Trigger create events after watcher started
	go func() {
		s.torrentFileWatcher.Wait()
		logrus.Debug("File watcher started, dispatching event for already present torrent files")
		for fullPath, info := range s.torrentFileWatcher.WatchedFiles() {
			s.torrentFileWatcher.Event <- watcher.Event{Op: watcher.Create, Path: fullPath, FileInfo: info}
		}
	}()

	go func() {
		logrus.Debug("Starting file watcher")
		if err := s.torrentFileWatcher.Start(s.fileWatcherPoll); err != nil {
			logrus.WithError(err).Error("File watcher has stopped with an error")
		}
	}()
	go func() {
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			defer logrus.Debug("Exiting file watcher loop")
		}
		for {
			select {
			case event := <-s.torrentFileWatcher.Event:
				switch event.Op {
				case watcher.Create:
					logrus.Info(event)
					e := s.onTorrentFileCreate(event.Path)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"file":  filepath.Base(event.Path),
							"event": event.Op,
						}).WithError(err).Error("Error in file creation callback")
						continue
					}
				case watcher.Rename:
					logrus.Info(event)
					e := s.onTorrentFileRenamed(event.OldPath, event.Path)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"file":  filepath.Base(event.Path),
							"event": event.Op,
						}).WithError(err).Error("Error in file rename callback")
						continue
					}
				case watcher.Remove:
					logrus.Info(event)
					e := s.onTorrentFileRemoved(event.Path)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"file":  filepath.Base(event.Path),
							"event": event.Op,
						}).WithError(err).Error("Error in file remove callback")
						continue
					}
				default:
					// does not handle WRITE since the write may occur while the file is being written before CREATE
					logrus.WithFields(logrus.Fields{
						"file":  filepath.Base(event.Path),
						"event": event.Op,
					}).WithError(err).Info("Event is ignored")
				}
			case err := <-s.torrentFileWatcher.Error:
				logrus.WithError(err).Error("File watcher has reported an error")
			case <-s.torrentFileWatcher.Closed:
				logrus.Info("File watcher stopped")
				return
			}
		}
	}()

	return nil
}

func (s *SeedManager) onTorrentFileCreate(filePath string) error {
	f, a := os.OpenFile(filePath, os.O_RDONLY|os.O_EXCL, 0)
	if a != nil {
		sleep := 5 * time.Second
		logrus.WithFields(logrus.Fields{
			"file": filePath,
		}).Warnf("File is already in use, wait %s before proceed", sleep)
		time.Sleep(sleep) // File was most likely created but not written yet, let's wait just a bit
	} else {
		_ = f.Close()
	}

	torrentSeed, err := seed.LoadFromFile(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to create torrent from file")
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if _, contains := s.seeds[*torrentSeed.InfoHash()]; !contains {
		s.seeds[*torrentSeed.InfoHash()] = torrentSeed
		go func() {
			defer func() {
				s.lock.Lock()
				delete(s.seeds, *torrentSeed.InfoHash())
				s.lock.Unlock()
			}()
			torrentSeed.Seed(s.client, s.bandwidthDispatcher)
		}()
	} else {
		logrus.WithFields(logrus.Fields{
			"file": filepath.Base(filePath),
		}).Warn("Seed was not not started, seed map already contains this infohash.")
	}

	return nil
}
func (s *SeedManager) onTorrentFileRenamed(oldFilePath string, newFilesPath string) error {
	s.lock.Lock()
	// Run the stop synchronously to ensure the STOP will be send before the START
	found := false
	for _, v := range s.seeds {
		filename := filepath.Base(oldFilePath)
		if filepath.Base(v.FilePath()) == filename {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			v.StopSeeding(ctx)
			found = true
			break
		}
	}
	s.lock.Unlock()
	if !found {
		return errors.New("cannot remove torrent '%s' from seeding list: not found in list")
	}

	return s.onTorrentFileCreate(newFilesPath)
}
func (s *SeedManager) onTorrentFileRemoved(filePath string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, v := range s.seeds {
		filename := filepath.Base(filePath)
		if filepath.Base(v.FilePath()) == filename {
			go func() {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				v.StopSeeding(ctx)
			}()
			return nil
		}
	}

	return errors.New("cannot remove torrent '%s' from seeding list: not found in list")
}

func (s *SeedManager) Stop(ctx context.Context) {
	logrus.Info("Stopping seedmanager gracefully")
	s.lock.Lock()
	defer s.lock.Unlock()

	logrus.WithFields(logrus.Fields{
		"seedCount": len(s.seeds),
	}).Info("Trigger seeds shutdown")
	wg := sync.WaitGroup{}
	for _, v := range s.seeds {
		wg.Add(1)
		go func() {
			v.StopSeeding(ctx)
			wg.Done()
		}()
	}

	s.client.StopListener(ctx)
	s.bandwidthDispatcher.Stop()
	s.torrentFileWatcher.Close()
	s.torrentFileWatcher = nil

	wg.Wait()
	s.seeds = make(map[torrent.InfoHash]seed.ISeed)
}
