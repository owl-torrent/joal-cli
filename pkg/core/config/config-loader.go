package config

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path/filepath"
)

const (
	torrentFolder          = "torrents"
	archivedTorrentFolders = torrentFolder + string(os.PathSeparator) + "archived"
	clientsFolder          = "clients"
	runtimeConfigFile      = "config.yml"
)

type IConfigLoader interface {
	LoadConfigAndInitIfNeeded() (*JoalConfig, error)
}

type joalConfigLoader struct {
	clientDownloader iClientDownloader
	configLocation   string
}

func NewJoalConfigLoader(configDir string, client *http.Client) (IConfigLoader, error) {
	return &joalConfigLoader{
		clientDownloader: newClientDownloader(filepath.Join(configDir, clientsFolder), client, newGithubClient(client)),
		configLocation:   configDir,
	}, nil
}

func (l *joalConfigLoader) LoadConfigAndInitIfNeeded() (*JoalConfig, error) {
	log := logs.GetLogger()
	err := applyMigrations()
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply migration step")
	}

	if hasInitialSetup, err := hasInitialSetup(l.configLocation); err != nil {
		return nil, err
	} else if !hasInitialSetup {
		if err := initialSetup(l.configLocation); err != nil {
			return nil, err
		}
	}

	if isInstalled, version, err := l.clientDownloader.IsInstalled(); err != nil {
	} else if !isInstalled {
		log.Info("config loader: client files are not installed, going to install them")
		err = l.clientDownloader.Install()
		if err != nil {
			return nil, err
		}
	} else if isInstalled {
		log.Info("config loader: client files are installed", zap.String("version", version.String()))
	}

	runtimeConfig := readRuntimeConfigOrDefault(filepath.Join(l.configLocation, runtimeConfigFile))

	conf := &JoalConfig{
		TorrentsDir:         filepath.Join(l.configLocation, torrentFolder),
		ArchivedTorrentsDir: filepath.Join(l.configLocation, archivedTorrentFolders),
		ClientsDir:          filepath.Join(l.configLocation, clientsFolder),
		RuntimeConfig:       runtimeConfig,
	}
	log.Info("config loader: config successfully loaded", zap.Any("config", conf))
	return conf, nil
}

func readRuntimeConfigOrDefault(filePath string) *RuntimeConfig {
	log := logs.GetLogger()
	runtimeConfig := RuntimeConfig{}.Default()

	f, err := os.Open(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("config loader: failed to open runtime config file '%s', running with default config instead", filePath), zap.Error(err))
		return runtimeConfig
	}
	defer func() { _ = f.Close() }()

	err = yaml.NewDecoder(f).Decode(runtimeConfig)
	if err != nil {
		log.Error(fmt.Sprintf("config loader: failed to parse runtime config file '%s', running with default config instead", filePath), zap.Error(err))
		return runtimeConfig
	}
	return runtimeConfig
}

// Check if all minimal required files are present on disk
func hasInitialSetup(rootConfigFolder string) (bool, error) {
	requiredPath := []string{
		rootConfigFolder,
		filepath.Join(rootConfigFolder, torrentFolder),
		filepath.Join(rootConfigFolder, archivedTorrentFolders),
		filepath.Join(rootConfigFolder, clientsFolder),
		filepath.Join(rootConfigFolder, runtimeConfigFile),
	}

	for _, dir := range requiredPath {
		_, err := os.Stat(dir)
		if err != nil && !os.IsNotExist(err) {
			return false, errors.Wrapf(err, "failed to read folder '%s'", dir)
		}
		if os.IsNotExist(err) {
			return false, nil
		}
	}

	return true, nil
}

// install all minimal required files to run
func initialSetup(rootConfigFolder string) error {
	requiredDirectories := []string{
		rootConfigFolder,
		filepath.Join(rootConfigFolder, torrentFolder),
		filepath.Join(rootConfigFolder, archivedTorrentFolders),
		filepath.Join(rootConfigFolder, clientsFolder),
	}

	for _, dir := range requiredDirectories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "failed to create folder '%s'", dir)
		}
	}

	// do not override config if already present
	if _, err := os.Stat(filepath.Join(rootConfigFolder, runtimeConfigFile)); err != nil {
		f, err := os.OpenFile(filepath.Join(rootConfigFolder, runtimeConfigFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return errors.Wrapf(err, "failed to open file '%s'", filepath.Join(rootConfigFolder, runtimeConfigFile))
		}
		defer func() { _ = f.Close() }()

		if err := yaml.NewEncoder(f).Encode(RuntimeConfig{}.Default()); err != nil {
			return errors.Wrapf(err, "failed to marshal RuntimeConfig into '%s'", filepath.Join(rootConfigFolder, runtimeConfigFile))
		}
	}

	return nil
}

func applyMigrations() error {
	// apply migration operation between version if needed
	// TODO: to be done
	return nil
}
