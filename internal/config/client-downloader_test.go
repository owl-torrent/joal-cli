package config

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type mockedGithubRepoService struct {
	release  *github.RepositoryRelease
	response *github.Response
	error
}

func (r *mockedGithubRepoService) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
	return r.release, r.response, r.error
}

func createTarGzMockedArchive(t *testing.T) io.Reader {
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	defer func() { _ = gw.Close() }()
	tw := tar.NewWriter(gw)
	defer func() { _ = tw.Close() }()

	fileContent := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	header := new(tar.Header)
	header.Name = "hello.yml"
	header.Size = int64(len(fileContent))
	header.Mode = int64(0755)
	header.ModTime = time.Now()
	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	// copy the file data to the tarball

	if _, err := tw.Write(fileContent); err != nil {
		t.Fatal(err)
	}

	_ = tw.Close()
	_ = gw.Close()

	return bytes.NewReader(b.Bytes())
}
func startHttpAssetDownloadServer(t *testing.T) (closeServer func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/my-asset", func(writer http.ResponseWriter, request *http.Request) {
		if _, err := io.Copy(writer, createTarGzMockedArchive(t)); err != nil {
			t.Fatal(err)
		}
	})
	listener, err := net.Listen("tcp", ":9876")
	if err != nil {
		t.Fatal(err)
	}
	server := http.Server{Handler: mux}
	go func() {
		err := server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return func() { _ = server.Close() }
}

func TestGithubClientDownloader_newClientDownloader_ShouldCreateClientDownloader(t *testing.T) {
	gc := &gitHubClient{Repositories: &mockedGithubRepoService{}}
	c := &http.Client{}
	d := newClientDownloader("/dev/null", c, gc)

	assert.Equal(t, "/dev/null", d.(*githubClientDownloader).clientsDirectory)
	assert.Equal(t, c, d.(*githubClientDownloader).httpClient)
	assert.Equal(t, gc, d.(*githubClientDownloader).githubClient)
	assert.Equal(t, clientFilesReleaseTag, d.(*githubClientDownloader).versionToInstall.String())
}

func TestGithubClientDownloader_IsInstalled_ShouldNotConsiderInstalledDirectoryDoesNotExists(t *testing.T) {
	d := newClientDownloader("/not/existing/path", &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{}})

	installed, err := d.IsInstalled()
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, installed)
}

func TestGithubClientDownloader_IsInstalled_ShouldNotConsiderInstalledIfVersionFileIMissing(t *testing.T) {
	dir := t.TempDir()
	d := newClientDownloader(dir, &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{}})

	installed, err := d.IsInstalled()
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, installed)
}

func TestGithubClientDownloader_IsInstalled_ShouldNotConsiderInstalledIfVersionIsDifferent(t *testing.T) {
	dir := t.TempDir()
	d := newClientDownloader(dir, &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{}})
	err := ioutil.WriteFile(filepath.Join(dir, clientVersionFileName), []byte("950.156.20"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	installed, err := d.IsInstalled()
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, installed)
}

func TestGithubClientDownloader_IsInstalled_ShouldConsiderInstalledIfVersionIsTheSame(t *testing.T) {
	dir := t.TempDir()
	d := newClientDownloader(dir, &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{}})
	err := ioutil.WriteFile(filepath.Join(dir, clientVersionFileName), []byte(clientFilesReleaseTag), 0755)
	if err != nil {
		t.Fatal(err)
	}

	installed, err := d.IsInstalled()
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, installed)
}

func TestGithubClientDownloader_Install_ShouldCreateOutputFolderIfMissing(t *testing.T) {
	assertDownloadUrl := "http://localhost:9876/my-asset"
	dir := t.TempDir()
	d := newClientDownloader(dir, &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{
		release: &github.RepositoryRelease{
			Assets: []*github.ReleaseAsset{
				{
					BrowserDownloadURL: &assertDownloadUrl,
				},
			},
		},
	}})

	err := syscall.Rmdir(dir)
	if err != nil {
		t.Fatal(err)
	}
	stopServer := startHttpAssetDownloadServer(t)
	defer stopServer()

	err = d.Install()
	if err != nil {
		t.Fatal(err)
	}

	assert.DirExists(t, dir)
}

func TestGithubClientDownloader_Install_ShouldFailIfGithubServiceReturnsReleaseWithMoreThanOneAsset(t *testing.T) {
	assertDownloadUrl := "http://localhost:9876/my-asset"
	dir := t.TempDir()
	d := newClientDownloader(dir, &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{
		release: &github.RepositoryRelease{
			Assets: []*github.ReleaseAsset{
				{BrowserDownloadURL: &assertDownloadUrl},
				{BrowserDownloadURL: &assertDownloadUrl},
			},
		},
	}})

	stopServer := startHttpAssetDownloadServer(t)
	defer stopServer()

	err := d.Install()
	if err != nil {
		assert.Contains(t, err.Error(), "to contains exactly one asset, asset contains 2")
	}
}

func TestGithubClientDownloader_Install_ShouldFailIfGithubServiceReturnsReleaseWithZeroAsset(t *testing.T) {
	dir := t.TempDir()
	d := newClientDownloader(dir, &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{
		release: &github.RepositoryRelease{
			Assets: []*github.ReleaseAsset{},
		},
	}})

	stopServer := startHttpAssetDownloadServer(t)
	defer stopServer()

	err := d.Install()
	if err != nil {
		assert.Contains(t, err.Error(), "to contains exactly one asset, asset contains 0")
	}
}

func TestGithubClientDownloader_Install_ShouldUnpackArchiveToOutputFolder(t *testing.T) {
	assertDownloadUrl := "http://localhost:9876/my-asset"
	dir := t.TempDir()
	d := newClientDownloader(dir, &http.Client{}, &gitHubClient{Repositories: &mockedGithubRepoService{
		release: &github.RepositoryRelease{
			Assets: []*github.ReleaseAsset{
				{
					BrowserDownloadURL: &assertDownloadUrl,
				},
			},
		},
	}})

	stopServer := startHttpAssetDownloadServer(t)
	defer stopServer()

	err := d.Install()
	if err != nil {
		t.Fatal(err)
	}

	assert.FileExists(t, filepath.Join(dir, "hello.yml"))
}
