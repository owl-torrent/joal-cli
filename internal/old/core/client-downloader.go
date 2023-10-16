package core

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/anthonyraymond/joal-cli/internal/old/core/logs"
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
	githubRepoOwner       = "owl-torrent"
	githubRepoName        = "owl-clients"
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
	release, _, err := d.githubClient.Repositories.GetReleaseByTag(context.Background(), githubRepoOwner, githubRepoName, clientFilesReleaseTag)
	if err != nil {
		return fmt.Errorf("client downloader: error when fetching release with tag '%s': %w", clientFilesReleaseTag, err)
	}

	if len(release.Assets) == 0 || len(release.Assets) > 1 {
		return fmt.Errorf("client downloader : expected release '%s' to contains exactly one asset, asset contains %d", release.GetTagName(), len(release.Assets))
	}
	asset := release.Assets[0]

	response, err := d.httpClient.Get(asset.GetBrowserDownloadURL())
	if err != nil {
		return fmt.Errorf("client downloader: failed to GET release from '%s': %w", asset.GetBrowserDownloadURL(), err)
	}

	if response.StatusCode >= 400 {
		_ = response.Body.Close()
		return fmt.Errorf("client downloader: failed to download release, response status code is %d: %w", response.StatusCode, err)
	}

	_, err = unpackit.Unpack(response.Body, d.clientsDirectory)
	if err != nil {
		return fmt.Errorf("client downloader: failed to unpack archive: %w", err)
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
		return nil, fmt.Errorf("failed to read client version file '%s': %w", filepath.Join(dir, clientVersionFileName), err)
	}

	version, err := semver.NewVersion(strings.TrimSpace(string(versionString)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse version '%s' as semvers from client version file '%s': %w", versionString, filepath.Join(dir, clientVersionFileName), err)
	}
	return version, nil
}
