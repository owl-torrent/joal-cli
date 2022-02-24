package core

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
)

var (
	configFileFromRoot = func(rootConfigDir string) string {
		return filepath.Join(rootConfigDir, "core.yml")
	}
	torrentDirFromRoot = func(rootConfigDir string) string {
		return filepath.Join(rootConfigDir, "torrents")
	}
	archivedTorrentDirFromRoot = func(rootConfigDir string) string {
		return filepath.Join(torrentDirFromRoot(rootConfigDir), "archived")
	}
	clientsDirFromRoot = func(rootConfigDir string) string {
		return filepath.Join(rootConfigDir, "clients")
	}
)

func Bootstrap(coreRootDir string, client *http.Client) (*CoreConfigLoader, error) {
	log := logs.GetLogger()

	err := os.MkdirAll(torrentDirFromRoot(coreRootDir), 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory '%s': %w", coreRootDir, err)
	}

	err = bootstrapConfigFile(coreRootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap core: %w", err)
	}
	err = bootstrapTorrentDirectories(coreRootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap core: %w", err)
	}

	configLoader := newCoreConfigLoader(coreRootDir)

	err = bootstrapClients(coreRootDir, client, log)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap core: %w", err)
	}

	return configLoader, nil
}

func bootstrapConfigFile(coreRootDir string) error {
	// Create the configuration file if missing
	f, err := os.OpenFile(configFileFromRoot(coreRootDir), os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("failed to create '%s' file: %w", configFileFromRoot(coreRootDir), err)
	}
	defer func() { _ = f.Close() }()
	return nil
}

func bootstrapTorrentDirectories(coreRootDir string) error {
	err := os.MkdirAll(torrentDirFromRoot(coreRootDir), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", torrentDirFromRoot(coreRootDir), err)
	}
	err = os.MkdirAll(archivedTorrentDirFromRoot(coreRootDir), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", archivedTorrentDirFromRoot(coreRootDir), err)
	}
	return nil
}

func bootstrapClients(coreRootDir string, client *http.Client, log *zap.Logger) error {
	log = log.With(zap.String("step", "clients"))
	clientDir := clientsDirFromRoot(coreRootDir)
	err := os.MkdirAll(clientDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", clientDir, err)
	}
	downloader := newClientDownloader(clientDir, client, newGithubClient(client))

	installed, version, err := downloader.IsInstalled()
	if err != nil {
		log.Warn("failed to check if clients files were installed, assume not installed")
	}
	if err == nil && installed {
		log.Info(fmt.Sprintf("clients files are installed, current version is %s", version.String()))
		return nil
	}
	err = downloader.Install()
	if err != nil {
		return fmt.Errorf("failed to install client files: %w", err)
	}

	return nil
}
