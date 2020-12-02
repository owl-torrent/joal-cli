package config

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	configLocation := configDir

	var err error
	if strings.TrimSpace(configLocation) == "" {
		configLocation, err = getDefaultConfigFolder()
		if err != nil {
			return nil, errors.Wrap(err, "config loader: failed to resolve default config folder")
		}
	}
	configLocation, err = filepath.Abs(configLocation)
	if err != nil {
		return nil, errors.Wrapf(err, "config loader: failed to transform '%s' to an absolute path", configLocation)
	}
	return &joalConfigLoader{
		clientDownloader: newClientDownloader(filepath.Join(configLocation, clientsFolder), newGithubClient(client)),
		configLocation:   configLocation,
	}, nil
}

func getDefaultConfigFolder() (string, error) {
	// Windows => %AppData%/joal
	// Mac     => $HOME/Library/Application Support/joal
	// Linux   => $XDG_CONFIG_HOME/joal or $HOME/.config/joal
	dir, err := os.UserConfigDir()
	return filepath.Join(dir, "joal"), err
}

func (l *joalConfigLoader) LoadConfigAndInitIfNeeded() (*JoalConfig, error) {
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

	if isInstalled, err := l.clientDownloader.IsInstalled(); err != nil {
		return nil, err
	} else if !isInstalled {
		err = l.clientDownloader.Install()
		if err != nil {
			return nil, err
		}
	}

	runtimeConfig := readRuntimeConfigOrDefault(filepath.Join(l.configLocation, runtimeConfigFile))

	return &JoalConfig{
		TorrentsDir:         filepath.Join(l.configLocation, torrentFolder),
		ArchivedTorrentsDir: filepath.Join(l.configLocation, archivedTorrentFolders),
		ClientsDir:          filepath.Join(l.configLocation, clientsFolder),
		RuntimeConfig:       runtimeConfig,
	}, nil
}

func readRuntimeConfigOrDefault(filePath string) *RuntimeConfig {
	log := logs.GetLogger()
	runtimeConfig := RuntimeConfig{}.Default()

	f, err := os.Open(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("config loader: failed to open runtime config file '%s', running with default config instead", filePath), zap.Error(err))
		return runtimeConfig
	}

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
	f, err := os.OpenFile(filepath.Join(rootConfigFolder, runtimeConfigFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to open file '%s'", filepath.Join(rootConfigFolder, runtimeConfigFile))
	}

	if err := yaml.NewEncoder(f).Encode(RuntimeConfig{}.Default()); err != nil {
		return errors.Wrapf(err, "failed to marshal RuntimeConfig into '%s'", filepath.Join(rootConfigFolder, runtimeConfigFile))
	}
	return nil
}

func applyMigrations() error {
	// apply migration operation between version if needed
	// TODO: to be done
	return nil
}
