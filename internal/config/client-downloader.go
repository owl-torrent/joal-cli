package config

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
	"github.com/c4milo/unpackit"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	clientVersionFileName = ".version"
	clientFilesReleaseTag = "1.0.0"
)

type githubRepoService interface {
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
}

type gitHubClient struct {
	Repositories githubRepoService
	// optionally store and export the underlying *github.Client
	// if you want easy access to client.Rate or other fields
}

func newGithubClient(httpClient *http.Client) *gitHubClient {
	client := github.NewClient(httpClient)

	return &gitHubClient{
		Repositories: client.Repositories,
	}
}

type iClientDownloader interface {
	IsInstalled() (bool, *semver.Version, error)
	Install() error
}

type githubClientDownloader struct {
	clientsDirectory string
	httpClient       *http.Client
	githubClient     *gitHubClient
	versionToInstall *semver.Version
}

func newClientDownloader(dest string, httpClient *http.Client, gitHubClient *gitHubClient) iClientDownloader {
	return &githubClientDownloader{
		clientsDirectory: dest,
		httpClient:       httpClient,
		githubClient:     gitHubClient,
		versionToInstall: semver.MustParse(clientFilesReleaseTag),
	}
}

func (d *githubClientDownloader) IsInstalled() (bool, *semver.Version, error) {
	log := logs.GetLogger()
	currentVersion, err := installedVersion(d.clientsDirectory)
	if err != nil {
		log.Info("client downloader: couldn't parse client version file, assume client are not installed", zap.NamedError("reason", err))
		return false, nil, nil
	}

	return currentVersion.Equal(d.versionToInstall), currentVersion, nil
}

func (d *githubClientDownloader) Install() error {
	release, _, err := d.githubClient.Repositories.GetReleaseByTag(context.Background(), "owl-torrent", "owl-clients", clientFilesReleaseTag)
	if err != nil {
		return errors.Wrapf(err, "client downloader: error when fetching release with tag '%s'", clientFilesReleaseTag)
	}

	if len(release.Assets) == 0 || len(release.Assets) > 1 {
		return fmt.Errorf("client downloader : expected release '%s' to contains exactly one asset, asset contains %d", release.GetTagName(), len(release.Assets))
	}
	asset := release.Assets[0]

	response, err := d.httpClient.Get(asset.GetBrowserDownloadURL())
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
	defer func() { _ = f.Close() }()

	versionString, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read client version file '%s'", filepath.Join(dir, clientVersionFileName))
	}

	version, err := semver.NewVersion(strings.TrimSpace(string(versionString)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse version '%s' as semvers from client version file '%s'", versionString, filepath.Join(dir, clientVersionFileName))
	}
	return version, nil
}
