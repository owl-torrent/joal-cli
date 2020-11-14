package seedmanager

/*
type SeedManager struct {
	conf                *config.SeedConfig
	joalPaths           *joalPaths
	seeds               map[torrent.InfoHash]seed.ISeed
	torrentFileWatcher  *watcher.Watcher
	bandwidthDispatcher bandwidth.IDispatcher
	client              emulatedclient.IEmulatedClient
	fileWatcherPoll     time.Duration
	lock                *sync.Mutex
}

func SeedManagerNew(joalPaths *joalPaths, conf config.SeedConfig) (*SeedManager, error) {
	dispatcher := bandwidth.dispatcherNew(&bandwidth.randomSpeedProvider{
		MinimumBytesPerSeconds: conf.MinUploadRate,
		MaximumBytesPerSeconds: conf.MaxUploadRate,
	})

	client, err := emulatedclient.FromClientFile(path.Join(joalPaths.clientFileFolder, conf.Client))
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
	panic("not implemented")
}

func (s *SeedManager) Start() error {
	log := logs.GetLogger()
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
		log.Debug("File watcher started, dispatching event for already present torrent files")
		for fullPath, info := range s.torrentFileWatcher.WatchedFiles() {
			s.torrentFileWatcher.Event <- watcher.Event{Op: watcher.Create, Path: fullPath, FileInfo: info}
		}
	}()

	go func() {
		log.Debug("Starting file watcher")
		if err := s.torrentFileWatcher.Start(s.fileWatcherPoll); err != nil {
			log.Error("Starting file watcher", zap.Error(err))
		}
	}()
	go func() {
		if log.Core().Enabled(zap.DebugLevel) {
			defer log.Debug("Exiting file watcher loop")
		}
		for {
			select {
			case event := <-s.torrentFileWatcher.Event:
				switch event.Op {
				case watcher.Create:
					log.Info(event.String())
					e := s.onTorrentFileCreate(event.Path)
					if e != nil {
						log.Error("Error in file creation callback",
							zap.String("file", filepath.Base(event.Path)),
							zap.Any("event", event.Op),
							zap.Error(err),
						)
						continue
					}
				case watcher.Rename:
					log.Info(event.String())
					e := s.onTorrentFileRenamed(event.OldPath, event.Path)
					if e != nil {
						log.Error("Error in file rename callback", zap.String("file", filepath.Base(event.Path)), zap.String("event", event.Op.String()))
						continue
					}
				case watcher.Remove:
					log.Info(event.String())
					e := s.onTorrentFileRemoved(event.Path)
					if e != nil {
						log.Error("Error in file remove callback", zap.String("file", filepath.Base(event.Path)), zap.String("event", event.Op.String()))
						continue
					}
				default:
					// does not handle WRITE since the write may occur while the file is being written before CREATE
					log.Info("Event is ignored", zap.String("file", filepath.Base(event.Path)), zap.String("event", event.Op.String()))
				}
			case err := <-s.torrentFileWatcher.Error:
				log.Error("File watcher has reported an error", zap.Error(err))
			case <-s.torrentFileWatcher.Closed:
				log.Info("File watcher stopped")
				return
			}
		}
	}()

	return nil
}

func (s *SeedManager) onTorrentFileCreate(filePath string) error {
	log := logs.GetLogger()
	f, a := os.OpenFile(filePath, os.O_RDONLY|os.O_EXCL, 0)
	if a != nil {
		sleep := 5 * time.Second
		if log.Core().Enabled(zap.WarnLevel) {
			log.Warn(fmt.Sprintf("File is already in use, wait %s before proceed", sleep), zap.String("file", filePath))
		}
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
			if _, contains := s.seeds[torrentSeed.InfoHash()]; !contains {
				s.seeds[torrentSeed.InfoHash()] = torrentSeed
				go func() {
					defer func() {
						s.lock.Lock()
						delete(s.seeds, torrentSeed.InfoHash())
						s.lock.Unlock()
					}()
					torrentSeed.Seed(s.client, s.bandwidthDispatcher)
				}()
			} else {
			log.Warn("Seed was not not started, seed map already contains this infohash.", zap.String("file", filepath.Base(filePath)))
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
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			v.StopSeeding(ctx)
			cancel()
			found = true
			break
		}
	}
	s.lock.Unlock()
	if !found {
		return fmt.Errorf("cannot remove torrent '%s' from seeding list: not found in list", oldFilePath)
	}

	return s.onTorrentFileCreate(newFilesPath)
}
func (s *SeedManager) onTorrentFileRemoved(filePath string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, v := range s.seeds {
		filename := filepath.Base(filePath)
		if filepath.Base(v.FilePath()) == filename {
			go func(s seed.ISeed) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				s.StopSeeding(ctx)
				cancel()
			}(v)
			return nil
		}
	}

	return fmt.Errorf("cannot remove torrent '%s' from seeding list: not found in list", filePath)
}

func (s *SeedManager) Stop(ctx context.Context) {
	log := logs.GetLogger()
	log.Info("Stopping seedmanager gracefully")
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Info("Trigger seeds shutdown", zap.Int("seedCount", len(s.seeds)))

	wg := sync.WaitGroup{}
	for _, v := range s.seeds {
		wg.Add(1)
		go func(s seed.ISeed) {
			defer wg.Done()
			s.StopSeeding(ctx)
		}(v)
	}

	s.client.StopListener(ctx)
	s.bandwidthDispatcher.Stop()
	s.torrentFileWatcher.Close()
	s.torrentFileWatcher = nil

	wg.Wait()
	s.seeds = make(map[torrent.InfoHash]seed.ISeed)
}
*/
