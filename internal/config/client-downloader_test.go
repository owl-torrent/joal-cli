package config

import (
	"context"
	"github.com/google/go-github/v32/github"
	"testing"
)

type mockedGithubRepoService struct {
	release  *github.RepositoryRelease
	response *github.Response
	error
}

func (r *mockedGithubRepoService) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
	return r.release, r.response, r.error
}

func TestGithubClientDownloader_newClientDownloader_ShouldCreateClientDownloader(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_IsInstalled_ShouldNotConsiderInstalledDirectoryDoesNotExists(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_IsInstalled_ShouldNotConsiderInstalledIfVersionFileIMissing(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_IsInstalled_ShouldNotConsiderInstalledIfVersionIsDifferent(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_IsInstalled_ShouldNotConsiderInstalledIfVersionIsTheSame(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_installedVersion_shouldParseVersionFromFile(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_Install_ShouldCreateOutputFolderIfMissing(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_Install_ShouldFailIfGithubServiceReturnsReleaseWithMoreThanOneAsset(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_Install_ShouldFailIfGithubServiceReturnsReleaseWithZeroAsset(t *testing.T) {
	t.Fatal("not implemented")
}

func TestGithubClientDownloader_Install_ShouldUnpackArchiveToOutputFolder(t *testing.T) {
	t.Fatal("not implemented")
}
