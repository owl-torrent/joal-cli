package core

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/old/common/configloader"
)

type CoreConfigLoader struct {
	configFilePath      string
	torrentDir          string
	archivedTorrentsDir string
	clientsDir          string
}

func newCoreConfigLoader(coreRootDir string) *CoreConfigLoader {
	return &CoreConfigLoader{
		configFilePath:      configFileFromRoot(coreRootDir),
		torrentDir:          torrentDirFromRoot(coreRootDir),
		archivedTorrentsDir: archivedTorrentDirFromRoot(coreRootDir),
		clientsDir:          clientsDirFromRoot(coreRootDir),
	}
}

func (l *CoreConfigLoader) ReadConfig() (*JoalConfig, error) {
	conf := RuntimeConfig{}.Default()
	err := configloader.ParseIntoDefault(l.configFilePath, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RuntimeConfig: %w", err)
	}

	return &JoalConfig{
		TorrentsDir:         l.torrentDir,
		ArchivedTorrentsDir: l.archivedTorrentsDir,
		ClientsDir:          l.clientsDir,
		RuntimeConfig:       conf,
	}, nil
}

// the boolean is true if joal needs to be restarted in order for the config to apply
func (l *CoreConfigLoader) SaveConfigToFile(newConf *RuntimeConfig) error {
	err := configloader.SaveToFile(l.configFilePath, newConf)
	if err != nil {
		return fmt.Errorf("failed to save RuntimeConfig: %w", err)
	}

	return nil
}
