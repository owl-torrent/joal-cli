package web

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
)

var (
	staticFilesDirFromRoot = func(pluginRootDir string) string {
		return filepath.Join(pluginRootDir, "web-resources")
	}
	webConfigFilePathFromRoot = func(pluginRootDir string) string {
		return filepath.Join(pluginRootDir, "web.yml")
	}
)

func bootstrap(configRoot string, client *http.Client, log *zap.Logger) error {
	if err := os.MkdirAll(configRoot, 0755); err != nil {
		return errors.Wrapf(err, "failed to create folder '%s'", configRoot)
	}

	f, err := os.OpenFile(webConfigFilePathFromRoot(configRoot), os.O_CREATE, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create '%s' file", webConfigFilePathFromRoot(configRoot))
	}
	_ = f.Close()

	err = bootstrapWebUi(configRoot, client, log)
	if err != nil {
		return errors.Wrap(err, "failed to download webui")
	}
	return nil
}

func bootstrapWebUi(coreRootDir string, client *http.Client, log *zap.Logger) error {
	log = log.With(zap.String("step", "clients"))
	clientDir := staticFilesDirFromRoot(coreRootDir)
	err := os.MkdirAll(clientDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create directory '%s'", clientDir)
	}
	downloader := newWebuiDownloader(clientDir, client, newGithubClient(client))

	installed, version, err := downloader.IsInstalled()
	if err != nil {
		log.Warn("failed to check if webui is installed, assume not installed")
	}
	if err == nil && installed {
		log.Info(fmt.Sprintf("webui is installed, current version is %s", version.String()))
		return nil
	}
	err = downloader.Install()
	if err != nil {
		return errors.Wrap(err, "failed to install webui files")
	}

	return nil
}
