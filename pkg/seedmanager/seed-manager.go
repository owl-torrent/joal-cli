package seedmanager

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anthonyraymond/joal-cli/internal/config"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients"
	"github.com/anthonyraymond/joal-cli/pkg/seed"
	"github.com/pkg/errors"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

type SeedConfig struct {
	MinUploadRate              int64
	MaxUploadRate              int64
	Client                     string
	RemoveTorrentWithZeroPeers bool
}

type seedManagerState int32

const (
	stopped seedManagerState = iota
	started
)

type joalPaths struct {
	torrentFolder        string
	torrentArchiveFolder string
	clientFileFolder     string
}

func joalPathsNew(joalWorkingDirectory string) (*joalPaths, error) {
	if !filepath.IsAbs(joalWorkingDirectory) {
		return nil, errors.New("joalWorkingDirectory must be an absolute path")
	}
	return &joalPaths{
		torrentFolder:        filepath.Join(joalWorkingDirectory, "torrents"),
		torrentArchiveFolder: filepath.Join(joalWorkingDirectory, "torrents", "archived"),
		clientFileFolder:     filepath.Join(joalWorkingDirectory, "clients"),
	}, nil
}

type SeedManager struct {
	state               seedManagerState
	configManager       *config.Manager
	joalPaths           *joalPaths
	seeds               map[torrent.InfoHash]*seed.Torrent
	torrentFileWatcher  *watcher.Watcher
	bandwidthDispatcher bandwidth.IDispatcher
	client              *emulatedclients.EmulatedClient
	lock                *sync.Mutex
	//TODO: eventListeners []EventListener // joal components will publish events from a chanel and seedmanager will relegate each of them in each of these publisher
}

func SeedManagerNew(joalWorkingDirectory string, configManager *config.Manager) (*SeedManager, error) {
	paths, err := joalPathsNew(joalWorkingDirectory)
	if err != nil {
		return nil, err
	}

	torrentFileWatcher := watcher.New()
	torrentFileWatcher.AddFilterHook(watcher.RegexFilterHook(regexp.MustCompile(`.+\.torrent$`), false))
	err = torrentFileWatcher.Add(paths.torrentFolder)
	if err != nil {
		return nil, err
	}

	return &SeedManager{
		state:               stopped,
		configManager:       configManager,
		joalPaths:           paths,
		seeds:               make(map[torrent.InfoHash]*seed.Torrent),
		torrentFileWatcher:  torrentFileWatcher,
		bandwidthDispatcher: nil,
		client:              nil,
		lock:                &sync.Mutex{},
	}, nil
}

func (s *SeedManager) Start() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.state == started {
		return errors.New("seedmanager is already started")
	}

	conf, err := s.configManager.Get()
	if err != nil {
		return err
	}
	s.bandwidthDispatcher = bandwidth.DispatcherNew(&bandwidth.RandomSpeedProvider{
		MinimumBytesPerSeconds: conf.MinUploadRate,
		MaximumBytesPerSeconds: conf.MaxUploadRate,
	})

	s.client, err = emulatedclients.FromClientFile(path.Join(s.joalPaths.clientFileFolder, conf.Client))
	if err != nil {
		return err
	}

	s.bandwidthDispatcher.Start()

	// Trigger create events after watcher started
	go func() {
		s.torrentFileWatcher.Wait()
		for _, f := range s.torrentFileWatcher.WatchedFiles() {
			s.torrentFileWatcher.TriggerEvent(watcher.Create, f)
		}
	}()

	if err := s.torrentFileWatcher.Start(1 * time.Second); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event := <-s.torrentFileWatcher.Event:
				fmt.Println(event) // Print the event's info.
				switch event.Op {
				case watcher.Create:
					e := s.onTorrentFileCreate(event.Path)
					if e != nil {
						//TODO: log error
						continue
					}
				case watcher.Rename, watcher.Write:
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

	s.state = started
	return nil
}

func (s *SeedManager) onTorrentFileCreate(filePath string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	torrentSeed, err := seed.LoadFromFile(filePath, s.client, s.bandwidthDispatcher)
	if err != nil {
		return errors.Wrap(err, "failed to create torrent from file")
	}

	if _, contains := s.seeds[*torrentSeed.InfoHash()]; contains {
		s.seeds[*torrentSeed.InfoHash()] = torrentSeed
		torrentSeed.Seed()
	}

	return nil
}
func (s *SeedManager) onTorrentFileRemoved(filePath string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	info, err := metainfo.LoadFromFile(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to create torrent metadata from file")
	}

	if v, contains := s.seeds[info.HashInfoBytes()]; contains {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		v.StopSeeding(ctx)
		delete(s.seeds, *v.InfoHash())
	}

	return nil
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
	if s.state != started {
		return
	}
	s.state = stopped

	wg := sync.WaitGroup{}
	for _, v := range s.seeds {
		wg.Add(1)
		go func() {
			v.StopSeeding(ctx)
			wg.Done()
		}()
	}

	s.bandwidthDispatcher.Stop()
	s.torrentFileWatcher.Close()

	wg.Wait()
	s.seeds = make(map[torrent.InfoHash]*seed.Torrent)
	s.bandwidthDispatcher = nil
	s.torrentFileWatcher = nil
}
