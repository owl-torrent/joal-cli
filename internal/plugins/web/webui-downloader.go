package web

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/c4milo/unpackit"
	"github.com/google/go-github/v42/github"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	githubRepoOwner      = "owl-torrent"
	githubRepoName       = "owl-webui"
	webuiVersionFileName = ".version"
	webuiFilesReleaseTag = "1.0.0"
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

type iWebuiDownloader interface {
	IsInstalled() (bool, *semver.Version, error)
	Install() error
}

type githubWebuiDownloader struct {
	webUiDirectory   string
	httpClient       *http.Client
	githubClient     *gitHubClient
	versionToInstall *semver.Version
}

func newWebuiDownloader(dest string, httpClient *http.Client, gitHubClient *gitHubClient) iWebuiDownloader {
	return &githubWebuiDownloader{
		webUiDirectory:   dest,
		httpClient:       httpClient,
		githubClient:     gitHubClient,
		versionToInstall: semver.MustParse(webuiFilesReleaseTag),
	}
}

func (d *githubWebuiDownloader) IsInstalled() (bool, *semver.Version, error) {
	log := logs.GetLogger()
	currentVersion, err := installedVersion(d.webUiDirectory)
	if err != nil {
		log.Info("webui downloader: couldn't parse webui version file, assume webui is not installed", zap.NamedError("reason", err))
		return false, nil, nil
	}

	return currentVersion.Equal(d.versionToInstall), currentVersion, nil
}

func (d *githubWebuiDownloader) Install() error {
	release, _, err := d.githubClient.Repositories.GetReleaseByTag(context.Background(), githubRepoOwner, githubRepoName, webuiFilesReleaseTag)
	if err != nil {
		return fmt.Errorf("webui downloader: error when fetching release with tag '%s': %w", webuiFilesReleaseTag, err)
	}

	if len(release.Assets) == 0 || len(release.Assets) > 1 {
		return fmt.Errorf("webui downloader : expected release '%s' to contains exactly one asset, asset contains %d", release.GetTagName(), len(release.Assets))
	}
	asset := release.Assets[0]

	response, err := d.httpClient.Get(asset.GetBrowserDownloadURL())
	if err != nil {
		return fmt.Errorf("webui downloader: failed to GET release from '%s': %w", asset.GetBrowserDownloadURL(), err)
	}

	if response.StatusCode >= 400 {
		_ = response.Body.Close()
		return fmt.Errorf("webui downloader: failed to download release, response status code is %d: %w", response.StatusCode, err)
	}

	_, err = unpackit.Unpack(response.Body, d.webUiDirectory)
	if err != nil {
		return fmt.Errorf("webui downloader: failed to unpack archive: %w", err)
	}

	return nil
}

func installedVersion(dir string) (*semver.Version, error) {
	f, err := os.Open(filepath.Join(dir, webuiVersionFileName))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	versionString, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read web version file '%s': %w", filepath.Join(dir, webuiVersionFileName), err)
	}

	version, err := semver.NewVersion(strings.TrimSpace(string(versionString)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse version '%s' as semvers from web version file '%s': %w", versionString, filepath.Join(dir, webuiVersionFileName), err)
	}
	return version, nil
}
