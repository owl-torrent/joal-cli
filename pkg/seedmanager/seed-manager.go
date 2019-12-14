package seedmanager

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients"
	"github.com/anthonyraymond/joal-cli/pkg/seed"
	"github.com/anthonyraymond/joal-cli/pkg/seedmanager/config"
	"github.com/pkg/errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

type SeedManager struct {
	conf                *config.SeedConfig
	joalPaths           *joalPaths
	seeds               map[torrent.InfoHash]*seed.Torrent
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
		seeds:               make(map[torrent.InfoHash]*seed.Torrent),
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
		for fullPath, info := range s.torrentFileWatcher.WatchedFiles() {
			if info.IsDir() { // TODO: remove this test when https://github.com/radovskyb/watcher/pull/88 gets merged and published
				continue
			}
			s.torrentFileWatcher.Event <- watcher.Event{Op: watcher.Create, Path: fullPath, FileInfo: info}
		}
	}()

	go func() {
		if err := s.torrentFileWatcher.Start(s.fileWatcherPoll); err != nil {
			// TODO: log error
		}
	}()
	go func() {
		for {
			select {
			case event := <-s.torrentFileWatcher.Event:
				if event.FileInfo.IsDir() { // TODO: remove this test when https://github.com/radovskyb/watcher/pull/88 gets merged and published
					continue
				}
				fmt.Println(event)
				switch event.Op {
				case watcher.Create:
					//TODO: logger.info(event) // Print the event's info.
					e := s.onTorrentFileCreate(event.Path)
					if e != nil {
						//TODO: log error
						continue
					}
				case watcher.Rename, watcher.Write:
					//TODO: logger.info(event) // Print the event's info.
					e := s.onTorrentFileRemoved(event.Path)
					if e != nil {
						//TODO: log error
						continue
					}
					e = s.onTorrentFileCreate(event.Path)
					if e != nil {
						//TODO: log error
						continue
					}
				case watcher.Remove:
					//TODO: logger.info(event) // Print the event's info.
					e := s.onTorrentFileRemoved(event.Path)
					if e != nil {
						//TODO: log error
						continue
					}
				default:
					// TODO: log action not handled
				}
			case err := <-s.torrentFileWatcher.Error:
				log.Fatalln(err)
			case <-s.torrentFileWatcher.Closed:
				return
			}
		}
	}()

	return nil
}

func (s *SeedManager) onTorrentFileCreate(filePath string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	torrentSeed, err := seed.LoadFromFile(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to create torrent from file")
	}

	if _, contains := s.seeds[*torrentSeed.InfoHash()]; !contains {
		s.seeds[*torrentSeed.InfoHash()] = torrentSeed
		torrentSeed.Seed(s.client, s.bandwidthDispatcher)
	}

	return nil
}
func (s *SeedManager) onTorrentFileRemoved(filePath string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, v := range s.seeds {
		filename := filepath.Base(filePath)
		if filepath.Base(v.FilePath()) == filename {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			v.StopSeeding(ctx)
			delete(s.seeds, *v.InfoHash())
			return nil
		}
	}

	return errors.New("cannot remove torrent '%s' from seeding list: not found in list")
}

func (s *SeedManager) onSeedStopHook(torrentSeed *seed.Torrent) seed.OnStopHook {
	return func() {
		// not lock to prevent dead lock when SeedManager.Stop() is called. This callback will be fired durnt the Stop() process
		delete(s.seeds, *torrentSeed.InfoHash())
	}
}

func (s *SeedManager) Stop(ctx context.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()

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
	s.seeds = make(map[torrent.InfoHash]*seed.Torrent)
}

func torrentFileFilter() watcher.FilterFileHookFunc {
	nameFilter := watcher.RegexFilterHook(regexp.MustCompile(`.+\.torrent$`), false)
	fileFilter := func(info os.FileInfo, fullPath string) error {
		if info.IsDir() {
			return watcher.ErrSkip
		}
		return nil
	}

	return func(info os.FileInfo, fullPath string) error {
		err := fileFilter(info, fullPath)
		if err != nil {
			return err
		}
		return nameFilter(info, fullPath)
	}
}
