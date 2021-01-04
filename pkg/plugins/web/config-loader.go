package web

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	staticFilesDir = "web-resources"
	webConfigFile  = "web.yml"
)

type IConfigLoader interface {
	LoadConfigAndInitIfNeeded() (*WebConfig, error)
}

type webConfigLoader struct {
	webuiDownloader iWebuiDownloader
	configLocation  string
}

func NewWebConfigLoader(configDir string, client *http.Client) (IConfigLoader, error) {
	configLocation := configDir

	var err error
	if strings.TrimSpace(configLocation) == "" {
		configLocation, err = getDefaultConfigFolder()
		if err != nil {
			return nil, errors.Wrap(err, "web config loader: failed to resolve default config folder")
		}
	}
	configLocation, err = filepath.Abs(configLocation)
	if err != nil {
		return nil, errors.Wrapf(err, "web config loader: failed to transform '%s' to an absolute path", configLocation)
	}
	return &webConfigLoader{
		webuiDownloader: newWebuiDownloader(filepath.Join(configLocation, staticFilesDir), client, newGithubClient(client)),
		configLocation:  configLocation,
	}, nil
}

func getDefaultConfigFolder() (string, error) {
	// Windows => %AppData%/joal
	// Mac     => $HOME/Library/Application Support/joal
	// Linux   => $XDG_CONFIG_HOME/joal or $HOME/.config/joal
	dir, err := os.UserConfigDir()
	return filepath.Join(dir, "joal", "web"), err
}

func (l *webConfigLoader) LoadConfigAndInitIfNeeded() (*WebConfig, error) {
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

	if isInstalled, version, err := l.webuiDownloader.IsInstalled(); err != nil {
	} else if !isInstalled {
		log.Info("web config loader: client files are not installed, going to install them")
		err = l.webuiDownloader.Install()
		if err != nil {
			return nil, err
		}
	} else if isInstalled {
		log.Info("web config loader: client files are installed", zap.String("version", version.String()))
	}

	conf := readWebConfigOrDefault(filepath.Join(l.configLocation, webConfigFile))

	log.Info("web config loader: config successfully loaded", zap.Any("config", conf))
	return conf, nil
}

func readWebConfigOrDefault(filePath string) *WebConfig {
	log := logs.GetLogger()
	webConfig := WebConfig{}.Default()

	f, err := os.Open(filePath)
	if err != nil {
		log.Error(fmt.Sprintf("web config loader: failed to open web config file '%s', running with default config instead", filePath), zap.Error(err))
		return webConfig
	}
	defer func() { _ = f.Close() }()

	err = yaml.NewDecoder(f).Decode(webConfig)
	if err != nil {
		log.Error(fmt.Sprintf("web config loader: failed to parse web config file '%s', running with default config instead", filePath), zap.Error(err))
		return webConfig
	}
	return webConfig
}

// Check if all minimal required files are present on disk
func hasInitialSetup(rootConfigFolder string) (bool, error) {
	requiredPath := []string{
		rootConfigFolder,
		filepath.Join(rootConfigFolder, staticFilesDir),
		filepath.Join(rootConfigFolder, staticFilesDir, "index.html"),
		filepath.Join(rootConfigFolder, webConfigFile),
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
		filepath.Join(rootConfigFolder, staticFilesDir),
	}

	for _, dir := range requiredDirectories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "failed to create folder '%s'", dir)
		}
	}

	// do not override config if already present
	if _, err := os.Stat(filepath.Join(rootConfigFolder, webConfigFile)); err != nil {
		f, err := os.OpenFile(filepath.Join(rootConfigFolder, webConfigFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return errors.Wrapf(err, "failed to open file '%s'", filepath.Join(rootConfigFolder, webConfigFile))
		}
		defer func() { _ = f.Close() }()

		if err := yaml.NewEncoder(f).Encode(WebConfig{}.Default()); err != nil {
			return errors.Wrapf(err, "failed to marshal WebConfig into '%s'", filepath.Join(rootConfigFolder, webConfigFile))
		}
	}

	return nil
}

func applyMigrations() error {
	// apply migration operation between version if needed
	// TODO: to be done
	return nil
}
