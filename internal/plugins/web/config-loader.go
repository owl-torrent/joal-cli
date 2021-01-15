package web

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path/filepath"
)

var (
	staticFilesDir = func(pluginRootDir string) string {
		return filepath.Join(pluginRootDir, "web-resources")
	}
	webConfigFilePath = func(pluginRootDir string) string {
		return filepath.Join(pluginRootDir, "web.yml")
	}
)

type IConfigLoader interface {
	LoadConfigAndInitIfNeeded() (*WebConfig, error)
}

type webConfigLoader struct {
	webuiDownloader iWebuiDownloader
	configLocation  string
	logger          *zap.Logger
}

func NewWebConfigLoader(configDir string, client *http.Client, logger *zap.Logger) (IConfigLoader, error) {
	return &webConfigLoader{
		webuiDownloader: newWebuiDownloader(staticFilesDir(configDir), client, newGithubClient(client)),
		configLocation:  configDir,
		logger:          logger,
	}, nil
}

func (l *webConfigLoader) LoadConfigAndInitIfNeeded() (*WebConfig, error) {
	log := l.logger
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

	if isInstalled, version, err := l.webuiDownloader.IsInstalled(); err != nil {
	} else if !isInstalled {
		log.Info("plugin is not installed or outdated")
		err = l.webuiDownloader.Install()
		if err != nil {
			return nil, err
		}
	} else if isInstalled {
		log.Info("plugin is already installed", zap.String("version", version.String()))
	}

	conf := l.readWebConfigOrDefault(webConfigFilePath(l.configLocation))

	log.Info("plugin config successfully loaded", zap.Any("config", conf))
	return conf, nil
}

func (l *webConfigLoader) readWebConfigOrDefault(filePath string) *WebConfig {
	log := l.logger
	webConfig := WebConfig{}.Default()

	f, err := os.Open(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("failed to open plugin config file '%s', running with default config instead", filePath), zap.Error(err))
		return webConfig
	}
	defer func() { _ = f.Close() }()

	err = yaml.NewDecoder(f).Decode(webConfig)
	if err != nil {
		log.Error(fmt.Sprintf("failed to parse plugin config file '%s', running with default config instead", filePath), zap.Error(err))
		return webConfig
	}
	return webConfig
}

// Check if all minimal required files are present on disk
func hasInitialSetup(rootConfigFolder string) (bool, error) {
	requiredPath := []string{
		rootConfigFolder,
		staticFilesDir(rootConfigFolder),
		filepath.Join(staticFilesDir(rootConfigFolder), "index.html"),
		webConfigFilePath(rootConfigFolder),
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
		staticFilesDir(rootConfigFolder),
	}

	for _, dir := range requiredDirectories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "failed to create folder '%s'", dir)
		}
	}

	configFilePath := webConfigFilePath(rootConfigFolder)
	// do not override config if already present
	if _, err := os.Stat(configFilePath); err != nil {
		f, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return errors.Wrapf(err, "failed to open file '%s'", filepath.Join(rootConfigFolder, configFilePath))
		}
		defer func() { _ = f.Close() }()

		if err := yaml.NewEncoder(f).Encode(WebConfig{}.Default()); err != nil {
			return errors.Wrapf(err, "failed to marshal WebConfig into '%s'", configFilePath)
		}
	}

	return nil
}

func applyMigrations() error {
	// apply migration operation between version if needed
	// TODO: to be done
	return nil
}
