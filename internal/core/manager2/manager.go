package manager2

import (
	"bytes"
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anthonyraymond/joal-cli/internal/core"
	"github.com/anthonyraymond/joal-cli/internal/core/announces"
	"github.com/anthonyraymond/joal-cli/internal/core/bandwidth"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/core/torrent2"
	"github.com/anthonyraymond/joal-cli/internal/utils/stop"
	"github.com/anthonyraymond/watcher"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var NoOpProxy = http.ProxyFromEnvironment

type Manager interface {
	StartSeeding()
	StopSeeding(ctx context.Context)
	SaveTorrentFile(filename string, bytes []byte)
	ArchiveTorrent(hash torrent.InfoHash)
	// Quit destroy the Manager in a non-recoverable way. To be called before exiting the program.
	Quit()
}

type managerImpl struct {
	isSeeding       bool
	commands        chan func()
	configLoader    *core.CoreConfigLoader
	loadedConfig    *core.JoalConfig
	announceQueue   *torrent2.AnnounceQueue
	speedDispatcher bandwidth.SpeedDispatcher
	client          emulatedclient.IEmulatedClient
	torrents        map[torrent.InfoHash]torrent2.Torrent
	quit            stop.Chan
}

func Run(configLoader *core.CoreConfigLoader) (Manager, error) {
	log := logs.GetLogger()
	m := &managerImpl{
		isSeeding:    false,
		commands:     make(chan func(), 50),
		configLoader: configLoader,
		torrents:     make(map[torrent.InfoHash]torrent2.Torrent),
		quit:         stop.NewChan(),
	}

	err := m.doReloadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to start Manager: %w", err)
	}

	m.speedDispatcher = bandwidth.NewSpeedDispatcher(m.loadedConfig.RuntimeConfig.BandwidthConfig)

	torrentFileWatcher := watcher.New()
	torrentFileWatcher.AddFilterHook(torrentFileFilter)
	_ = torrentFileWatcher.Add(m.loadedConfig.TorrentsDir)

	const intervalBetweenTorrentStatsUpdate = 15 * time.Second
	go func(m *managerImpl) {
		refreshStatsTicker := time.NewTimer(intervalBetweenTorrentStatsUpdate)
		for {
			select {
			case command := <-m.commands:
				command()
			case <-refreshStatsTicker.C:
				if !m.isSeeding {
					continue
				}
				for key, _ := range m.torrents {
					m.torrents[key].AddDataFor(intervalBetweenTorrentStatsUpdate)
				}
			case err := <-torrentFileWatcher.Error:
				log.Warn("file watcher has reported an error", zap.Error(err))
			case event := <-torrentFileWatcher.Event:
				switch event.Op {
				case watcher.Create:
					log.Info(event.String())
					t, err := torrent2.FromFile(event.Path)
					if err != nil {
						log.Error("failed to parse torrent from file", zap.Error(err))
						break
					}
					m.torrents[t.InfoHash()] = t
					if m.isSeeding {
						clientAbilities := m.client.GetAnnounceCapabilities()
						t.Start(torrent2.AnnounceProps{
							SupportHttpAnnounce:   m.client.SupportsHttpAnnounce(),
							SupportUdpAnnounce:    m.client.SupportsUdpAnnounce(),
							SupportAnnounceList:   clientAbilities.SupportAnnounceList,
							AnnounceToAllTiers:    clientAbilities.AnnounceToAllTiers,
							AnnounceToAllTrackers: clientAbilities.AnnounceToAllTrackersInTier,
						}, m.announceQueue, m.speedDispatcher) // On passe le dispatcher pour que le torrent puisse se register
					}
				case watcher.Rename:
					log.Info(event.String())
					t, found := findTorrent(m.torrents, event.OldPath)
					if !found {
						break
					}
					t.ChangePath(event.Path)
				case watcher.Remove:
					log.Info(event.String())
					t, found := findTorrent(m.torrents, event.OldPath)
					if !found {
						break
					}
					delete(m.torrents, t.InfoHash())
					if m.isSeeding {
						ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
						t.Stop(ctx)
						cancel()
					}
				default:
					// does not handle WRITE since it may occur while the file is being written before CREATE
					log.Info("Event is ignored", zap.String("file", filepath.Base(event.Path)), zap.String("event", event.Op.String()))
				}
			case stopRequest := <-m.quit:
				//goland:noinspection GoDeferInLoop
				defer func() {
					stopRequest.NotifyDone()
				}()

				// close & drain refreshStatsTicker
				refreshStatsTicker.Stop()
				select {
				case <-refreshStatsTicker.C:
				default:
				}

				torrentFileWatcher.Close()
				<-torrentFileWatcher.Closed
				m.doStopSeeding(stopRequest.Ctx())

				return
			}
		}
	}(m)

	// Trigger create events after watcher started (to take into account already present torrent files on startup)
	go func() {
		torrentFileWatcher.Wait()
		log.Info("file watcher: started", zap.String("monitored-folder", m.loadedConfig.TorrentsDir))
		for fullPath, info := range torrentFileWatcher.WatchedFiles() {
			torrentFileWatcher.Event <- watcher.Event{Op: watcher.Create, Path: fullPath, FileInfo: info}
		}
	}()

	go func() {
		if err := torrentFileWatcher.Start(2 * time.Second); err != nil {
			log.Error("failed to run file watcher", zap.Error(err))
		}
	}()

	return m, nil
}

func (m *managerImpl) doStartSeeding() error {
	if m.isSeeding {
		return fmt.Errorf("manager is already seeding")
	}

	if m.loadedConfig.RuntimeConfig.Client == "" {
		return fmt.Errorf("core config does not contains a client file, please take a look at the documentation")
	}
	client, err := emulatedclient.FromClientFile(filepath.Join(m.loadedConfig.ClientsDir, m.loadedConfig.RuntimeConfig.Client), NoOpProxy)
	if err != nil {
		return fmt.Errorf("failed to load client file: %w", err)
	}
	err = client.StartListener(NoOpProxy)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	m.client = client
	m.announceQueue = torrent2.NewAnnounceQueue()
	go RunQueueConsumer(m.announceQueue, func(request *announces.AnnounceRequest) {
		client.Announce(request)
	})

	m.isSeeding = true
	m.speedDispatcher.Start()

	return nil
}

func (m *managerImpl) doSaveTorrentFile(filename string, content []byte) error {
	meta, err := metainfo.Load(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to parse torrent file: %w", err)
	}

	filename = filepath.Join(m.loadedConfig.TorrentsDir, filename)
	w, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("failed to open file '%s' for writing: %w", filename, err)
	}

	err = meta.Write(w)
	if err != nil {
		return fmt.Errorf("failed to write to file '%s': %w", filename, err)
	}
	return nil
}

func (m *managerImpl) doArchiveTorrent(hash torrent.InfoHash) error {
	var torrentToRemove torrent2.Torrent = nil
	for _, t := range m.torrents {
		if t.InfoHash().HexString() == hash.HexString() {
			torrentToRemove = t
			break
		}
		return fmt.Errorf("torrent not found in seeding list")
	}

	err := torrentToRemove.MoveTo(m.loadedConfig.ArchivedTorrentsDir)
	if err != nil {
		return fmt.Errorf("failed to move torrent file to archive directory: %w", err)
	}

	return nil
}

func (m *managerImpl) doStopSeeding(ctx context.Context) {
	if !m.isSeeding {
		return
	}
	for _, t := range m.torrents {
		ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
		t.Stop(ctx)
		cancel()
	}
	m.speedDispatcher.Stop()
	m.announceQueue.DiscardFutureEnqueueAndDestroy()
	m.isSeeding = false
	m.client.StopListener(context.Background())
}

func (m *managerImpl) doReloadConfig() error {
	conf, err := m.configLoader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	m.loadedConfig = conf
	// TODO: Based on what have changed, maybe we can publish an event "restart required" to warn the user that a
	//  restart is needed to fully apply the new configuration (for example: a change of the RuntieConfig.Client need a restart)

	if m.speedDispatcher != nil {
		m.speedDispatcher.ReplaceSpeedConfig(m.loadedConfig.RuntimeConfig.BandwidthConfig.Speed)
	}
	return nil
}

func (m *managerImpl) StartSeeding() {
	log := logs.GetLogger()
	m.commands <- func() {
		err := m.doStartSeeding()
		if err != nil {
			log.Error("manager failed to start seeding", zap.Error(err))
			//TODO: find a way to return error?
		}
	}
}

func (m *managerImpl) SaveTorrentFile(filename string, bytes []byte) {
	log := logs.GetLogger()
	m.commands <- func() {
		err := m.doSaveTorrentFile(filename, bytes)
		if err != nil {
			log.Error("manager failed to start seeding", zap.Error(err))
			//TODO: find a way to return error?
		}
	}
}

func (m *managerImpl) ArchiveTorrent(hash torrent.InfoHash) {
	log := logs.GetLogger()
	m.commands <- func() {
		err := m.doArchiveTorrent(hash)
		if err != nil {
			log.Error("manager failed to archive torrent", zap.Error(err))
			//TODO: find a way to return error?
		}
	}
}

func (m *managerImpl) StopSeeding(ctx context.Context) {
	m.commands <- func() {
		m.doStopSeeding(ctx)
	}
}

func (m *managerImpl) Quit() {
	stopReq := stop.NewRequest(context.Background())
	m.quit <- stopReq

	_ = stopReq.AwaitDone()
}

func findTorrent(torrents map[torrent.InfoHash]torrent2.Torrent, path string) (torrent2.Torrent, bool) {
	for _, t := range torrents {
		if t.Path() == path {
			return t, true
		}
	}
	return nil, false
}
