package config

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/c4milo/unpackit"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	clientVersionFileName = ".version"
	clientFilesReleaseTag = "1.0.0"
)

type iClientDownloader interface {
	IsInstalled() (bool, error)
	Install() error
}

type githubClientDownloader struct {
	clientsDirectory string
	githubClient     *github.Client
	versionToInstall *semver.Version
}

func newClientDownloader(dest string) iClientDownloader {
	return &githubClientDownloader{
		clientsDirectory: dest,
		githubClient:     github.NewClient(nil),
		versionToInstall: semver.MustParse(clientFilesReleaseTag),
	}
}

func (d *githubClientDownloader) IsInstalled() (bool, error) {
	currentVersion, err := installedVersion(d.clientsDirectory)
	if err != nil {
		return false, nil
	}

	return currentVersion.Equal(d.versionToInstall), nil
}

func (d *githubClientDownloader) Install() error {
	release, _, err := d.githubClient.Repositories.GetReleaseByTag(context.Background(), "joal-torrent", "joal-clients", clientFilesReleaseTag)
	if err != nil {
		return errors.Wrapf(err, "client downloader: error when fetching release with tag '%s'", clientFilesReleaseTag)
	}

	if len(release.Assets) == 0 || len(release.Assets) > 1 {
		return fmt.Errorf("client downloader : expected release '%s' to contains exactly one asset, asset contains %d", release.GetTagName(), len(release.Assets))
	}
	if len(release.Assets) > 1 {
		return fmt.Errorf("client downloader : client release '%s' has more than one asset", release.GetTagName())
	}
	asset := release.Assets[0]

	response, err := http.Get(asset.GetBrowserDownloadURL())
	if err != nil {
		return errors.Wrapf(err, "client downloader: failed to GET release from '%s'", asset.GetBrowserDownloadURL())
	}

	if response.StatusCode >= 400 {
		_ = response.Body.Close()
		return errors.Wrapf(err, "client downloader: failed to download release, response status code is %d", response.StatusCode)
	}

	_, err = unpackit.Unpack(response.Body, d.clientsDirectory)
	if err != nil {
		return errors.Wrap(err, "client downloader: failed to unpack archive")
	}

	return nil
}

func installedVersion(dir string) (*semver.Version, error) {
	f, err := os.Open(filepath.Join(dir, clientVersionFileName))
	if err != nil {
		return nil, err
	}

	versionString, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return semver.NewVersion(string(versionString))
}
