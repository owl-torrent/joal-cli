package core

import (
	"github.com/anthonyraymond/joal-cli/internal/common/configloader"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "failed to parse RuntimeConfig")
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
		return errors.Wrap(err, "failed to save RuntimeConfig")
	}

	return nil
}
