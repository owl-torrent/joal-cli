package seedmanager

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/core/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/core/config"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/anthonyraymond/joal-cli/pkg/core/seed"
	"github.com/anthonyraymond/watcher"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// prend un objet JoalConfig (une struct qui contient la liste des fichiers clients (et leurs path), la liste des torrents, la liste des torrents archivés, la runtime config etc...

// gere les mouvement dans les dossiers

type ITorrentManager interface {
	Seed() error
	StopSeeding(ctx context.Context)
}

type torrentManager struct {
	lock         *sync.Mutex
	isRunning    bool
	configLoader config.IConfigLoader
	stopping     chan *stoppingRequest
}

func NewTorrentManager(configDir string) (ITorrentManager, error) {
	loader, err := config.NewJoalConfigLoader(configDir, &http.Client{})
	if err != nil {
		return nil, err
	}
	return &torrentManager{
		lock:         &sync.Mutex{},
		isRunning:    false,
		configLoader: loader,
		stopping:     make(chan *stoppingRequest),
	}, nil
}

type stoppingRequest struct {
	ctx          context.Context
	doneStopping chan struct{}
}

func (t *torrentManager) Seed() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.isRunning {
		return nil
	}
	t.isRunning = true

	log := logs.GetLogger()
	// Now that i used it, i feel like the configLoader should not init the config folder structure. It should be another part of the program that handles that.
	// Also the config loader should have been given to the TorrentManager as constructor argument and not build in constructor
	conf, err := t.configLoader.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.isRunning = false
		return errors.Wrap(err, "failed to load config")
	}

	client, err := emulatedclient.FromClientFile(filepath.Join(conf.ClientsDir, conf.RuntimeConfig.Client))
	if err != nil {
		t.isRunning = false
		return errors.Wrap(err, "failed to load client")
	}
	err = client.StartListener()
	if err != nil {
		t.isRunning = false
		return errors.Wrap(err, "failed to start listener")
	}

	torrents := make(map[string]seed.ITorrent)
	torrentFileWatcher := watcher.New()
	torrentFileWatcher.AddFilterHook(torrentFileFilter())
	_ = torrentFileWatcher.Add(conf.TorrentsDir)

	go func() {
		claimerPool := bandwidth.NewWeightedClaimerPool()
		dispatcher := bandwidth.NewDispatcher(conf.RuntimeConfig.BandwidthConfig.Dispatcher, claimerPool, bandwidth.NewRandomSpeedProvider(conf.RuntimeConfig.BandwidthConfig.Speed))
		dispatcher.Start()
		log.Info("torrent manager: started")

		for {
			select {
			case event := <-torrentFileWatcher.Event:
				switch event.Op {
				case watcher.Create:
					log.Info(event.String())
					t, err := seed.FromFile(event.Path)
					if err != nil {
						log.Error("failed to parse torrent from file", zap.Error(err))
						break
					}
					err = t.StartSeeding(client, claimerPool)
					if err != nil {
						log.Error("failed to start seeding", zap.Error(err))
						break
					}
					torrents[event.Path] = t
				case watcher.Rename:
					log.Info(event.String())
					t, ok := torrents[event.OldPath]
					if !ok {
						break
					}
					delete(torrents, event.OldPath)
					torrents[event.Path] = t
				case watcher.Remove:
					log.Info(event.String())
					t, ok := torrents[event.OldPath]
					if !ok {
						break
					}
					delete(torrents, event.Path)
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						t.StopSeeding(ctx)
					}()
				default:
					// does not handle WRITE since the write may occur while the file is being written before CREATE
					log.Info("Event is ignored", zap.String("file", filepath.Base(event.Path)), zap.String("event", event.Op.String()))
				}
			case err := <-torrentFileWatcher.Error:
				log.Warn("file watcher has reported an error", zap.Error(err))
			case stopRequest := <-t.stopping:
				torrentFileWatcher.Close()
				<-torrentFileWatcher.Closed

				wg := &sync.WaitGroup{}
				for _, t := range torrents {
					wg.Add(1)
					go func(t seed.ITorrent) {
						t.StopSeeding(stopRequest.ctx)
						wg.Done()
					}(t)
				}

				client.StopListener(stopRequest.ctx)
				dispatcher.Stop()

				doneAnnouncingStop := make(chan struct{})
				go func() {
					wg.Wait()
					close(doneAnnouncingStop)
				}()
				select {
				case <-doneAnnouncingStop:
				case <-stopRequest.ctx.Done():
				}

				stopRequest.doneStopping <- struct{}{}
				return
			}
		}
	}()

	// Trigger create events after watcher started (to take into account already present torrent files on startup)
	go func() {
		torrentFileWatcher.Wait()
		log.Info("file watcher: started", zap.String("monitored-folder", conf.TorrentsDir))
		for fullPath, info := range torrentFileWatcher.WatchedFiles() {
			torrentFileWatcher.Event <- watcher.Event{Op: watcher.Create, Path: fullPath, FileInfo: info}
		}
	}()

	go func() {
		if err := torrentFileWatcher.Start(1 * time.Second); err != nil {
			log.Error("failed to run file watcher", zap.Error(err))
		}
	}()

	return nil
}

func (t *torrentManager) StopSeeding(ctx context.Context) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.isRunning {
		return
	}
	t.isRunning = false
	log := logs.GetLogger()

	stopRequest := &stoppingRequest{
		ctx:          ctx,
		doneStopping: make(chan struct{}),
	}
	log.Info("torrent manager: stopping")
	t.stopping <- stopRequest

	<-stopRequest.doneStopping
	log.Info("torrent manager: stopped")
}
