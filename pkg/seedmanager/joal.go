package seedmanager

import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/seedmanager/config"
	"path/filepath"
	"sync"
)

type joalState int32

const (
	stopped joalState = iota
	started
)

type joalPaths struct {
	torrentFolder        string
	torrentArchiveFolder string
	clientFileFolder     string
}

func joalPathsNew(joalWorkingDirectory string) (*joalPaths, error) {
	if !filepath.IsAbs(joalWorkingDirectory) {
		return nil, fmt.Errorf("joalWorkingDirectory must be an absolute path")
	}
	return &joalPaths{
		torrentFolder:        filepath.Join(joalWorkingDirectory, "torrents"),
		torrentArchiveFolder: filepath.Join(joalWorkingDirectory, "torrents", "archived"),
		clientFileFolder:     filepath.Join(joalWorkingDirectory, "clients"),
	}, nil
}

type Joal struct {
	state         joalState
	configManager config.IManager
	joalPaths     *joalPaths
	seedManager   *SeedManager
	lock          *sync.Mutex
	//TODO: eventListeners []EventListener // joal components will publish events from a chanel and seedmanager will relegate each of them in each of these publisher
}

func JoalNew(joalWorkingDirectory string, configManager config.IManager) (*Joal, error) {
	paths, err := joalPathsNew(joalWorkingDirectory)
	if err != nil {
		return nil, err
	}

	return &Joal{
		state:         stopped,
		configManager: configManager,
		joalPaths:     paths,
		lock:          &sync.Mutex{},
	}, nil
}

func (j *Joal) Start() error {
	j.lock.Lock()
	defer j.lock.Unlock()
	if j.state == started {
		return fmt.Errorf("joal is already seeding")
	}

	conf, err := j.configManager.Get()
	if err != nil {
		return err
	}

	manager, err := SeedManagerNew(j.joalPaths, conf)
	if err != nil {
		return err
	}

	j.seedManager = manager
	err = manager.Start()
	if err != nil {
		return err
	}

	j.state = started
	return nil
}

func (j *Joal) Stop(ctx context.Context) {
	j.lock.Lock()
	defer j.lock.Unlock()
	if j.state != started {
		return
	}

	if j.seedManager != nil {
		j.seedManager.Stop(ctx)
	}
	j.seedManager = nil

	j.state = stopped
}
