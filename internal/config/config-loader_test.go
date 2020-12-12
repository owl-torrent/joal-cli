package config

import (
	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

type mockedClientDownloader struct {
	isInstalled func() (bool, *semver.Version, error)
	install     func() error
}

func (d *mockedClientDownloader) IsInstalled() (bool, *semver.Version, error) {
	if d.isInstalled != nil {
		return d.isInstalled()
	}
	return true, semver.MustParse("1.0.0"), nil
}

func (d *mockedClientDownloader) Install() error {
	if d.install != nil {
		return d.install()
	}
	return nil
}

func TestNewJoalConfigLoader_ShouldCreateNewConfigLoader(t *testing.T) {
	dir := t.TempDir()
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, l.(*joalConfigLoader).configLocation, dir)
	assert.NotNil(t, l.(*joalConfigLoader).clientDownloader)
}

func TestNewJoalConfigLoader_ShouldCreateNewConfigLoaderWithDefaultConfigPath(t *testing.T) {
	l, err := NewJoalConfigLoader("", &http.Client{})
	if err != nil {
		t.Fatal(err)
	}

	osDefault, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, l.(*joalConfigLoader).configLocation, osDefault)
}

func TestInitialSetup_ShouldCreateAllRequiredAndConfigFileFoldersIfMissing(t *testing.T) {
	dir := t.TempDir()
	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}

	assert.DirExists(t, filepath.Join(dir, clientsFolder))
	assert.DirExists(t, filepath.Join(dir, torrentFolder))
	assert.DirExists(t, filepath.Join(dir, archivedTorrentFolders))
	assert.FileExists(t, filepath.Join(dir, runtimeConfigFile))
}

func TestInitialSetup_ShouldCreateAllRequiredAndConfigFileFoldersIfMissingButDontOverrideConfigFileIfAlreadyPresent(t *testing.T) {
	dir := t.TempDir()

	if err := ioutil.WriteFile(filepath.Join(dir, runtimeConfigFile), []byte{61, 62}, 0755); err != nil {
		t.Fatal(err)
	}

	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}

	fileContent, err := ioutil.ReadFile(filepath.Join(dir, runtimeConfigFile))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []byte{61, 62}, fileContent)
}

func TestHasInitialSetup_ShouldFailIfMissingTorrentFolder(t *testing.T) {
	dir := t.TempDir()
	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Join(dir, torrentFolder)); err != nil {
		t.Fatal(err)
	}

	hasSetup, err := hasInitialSetup(dir)
	assert.NoError(t, err)
	assert.False(t, hasSetup)
}

func TestHasInitialSetup_ShouldFailIfMissingArchiveTorrentFolder(t *testing.T) {
	dir := t.TempDir()
	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Join(dir, archivedTorrentFolders)); err != nil {
		t.Fatal(err)
	}

	hasSetup, err := hasInitialSetup(dir)
	assert.NoError(t, err)
	assert.False(t, hasSetup)
}

func TestHasInitialSetup_ShouldFailIfMissingClientFolder(t *testing.T) {
	dir := t.TempDir()
	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Join(dir, clientsFolder)); err != nil {
		t.Fatal(err)
	}

	hasSetup, err := hasInitialSetup(dir)
	assert.NoError(t, err)
	assert.False(t, hasSetup)
}

func TestHasInitialSetup_ShouldFailIfMissingRuntimeConfig(t *testing.T) {
	dir := t.TempDir()
	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Join(dir, runtimeConfigFile)); err != nil {
		t.Fatal(err)
	}

	hasSetup, err := hasInitialSetup(dir)
	assert.NoError(t, err)
	assert.False(t, hasSetup)
}

func TestHasInitialSetup_ShouldSucceedIfInitialized(t *testing.T) {
	dir := t.TempDir()
	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}

	hasSetup, err := hasInitialSetup(dir)
	assert.NoError(t, err)
	assert.True(t, hasSetup)
}

func TestReadRuntimeConfigOrDefault_ShouldReturnDefaultConfigIfFolderDoesNotExists(t *testing.T) {
	dir := t.TempDir()
	if err := syscall.Rmdir(dir); err != nil {
		t.Fatal(err)
	}
	config := readRuntimeConfigOrDefault(filepath.Join(dir, runtimeConfigFile))

	assert.Equal(t, RuntimeConfig{}.Default(), config)
}

func TestReadRuntimeConfigOrDefault_ShouldReturnDefaultConfigIfFileDoesNotExists(t *testing.T) {
	dir := t.TempDir()
	config := readRuntimeConfigOrDefault(filepath.Join(dir, runtimeConfigFile))

	assert.Equal(t, config, RuntimeConfig{}.Default())
}

func TestReadRuntimeConfigOrDefault_ShouldReturnDefaultConfigIfFailsToParse(t *testing.T) {
	dir := t.TempDir()
	if err := ioutil.WriteFile(filepath.Join(dir, runtimeConfigFile), []byte("nop, thats not yaml"), 0755); err != nil {
		t.Fatal(err)
	}
	config := readRuntimeConfigOrDefault(filepath.Join(dir, runtimeConfigFile))

	assert.Equal(t, RuntimeConfig{}.Default(), config)
}

func TestReadRuntimeConfigOrDefault_ShouldReturnParseConfigIfFileIsPresent(t *testing.T) {
	dir := t.TempDir()
	runtimeConfig := RuntimeConfig{}.Default()
	runtimeConfig.Client = "HELLO :D"
	marshalled, err := yaml.Marshal(runtimeConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, runtimeConfigFile), marshalled, 0755); err != nil {
		t.Fatal(err)
	}
	config := readRuntimeConfigOrDefault(filepath.Join(dir, runtimeConfigFile))

	assert.Equal(t, runtimeConfig, config)
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldCreateFolderStructureIfNotPresent(t *testing.T) {
	dir := t.TempDir()
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	l.(*joalConfigLoader).clientDownloader = &mockedClientDownloader{}
	_, err = l.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.Fatal(err)
	}

	hasSetup, err := hasInitialSetup(dir)
	assert.NoError(t, err)
	assert.True(t, hasSetup)
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldAddMissingFolderStructureIfSomeFoldersAreMissing(t *testing.T) {
	dir := t.TempDir()
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	l.(*joalConfigLoader).clientDownloader = &mockedClientDownloader{}

	if err := initialSetup(dir); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Rmdir(filepath.Join(dir, clientsFolder)); err != nil {
		t.Fatal(err)
	}

	_, err = l.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.Fatal(err)
	}

	hasSetup, err := hasInitialSetup(dir)
	assert.NoError(t, err)
	assert.True(t, hasSetup)
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldDownloadClientsIfNotInstalled(t *testing.T) {
	dir := t.TempDir()
	hasInstalledClients := false
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	l.(*joalConfigLoader).clientDownloader = &mockedClientDownloader{
		isInstalled: func() (bool, *semver.Version, error) {
			return false, nil, nil
		},
		install: func() error {
			hasInstalledClients = true
			return nil
		},
	}

	_, err = l.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, hasInstalledClients)
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldNotDownloadClientsIfAlreadyInstalled(t *testing.T) {
	dir := t.TempDir()
	hasInstalledClients := false
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	l.(*joalConfigLoader).clientDownloader = &mockedClientDownloader{
		isInstalled: func() (bool, *semver.Version, error) {
			return true, semver.MustParse("1.0.0"), nil
		},
		install: func() error {
			hasInstalledClients = true
			return nil
		},
	}

	_, err = l.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, hasInstalledClients)
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldReturnDefaultConfigIfRuntimeConfigFileIsNotPresent(t *testing.T) {
	dir := t.TempDir()
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	l.(*joalConfigLoader).clientDownloader = &mockedClientDownloader{}
	if err != nil {
		t.Fatal(err)
	}

	config, err := l.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, RuntimeConfig{}.Default(), config.RuntimeConfig)
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldReturnParsedRuntimeConfigIfFileIsPresent(t *testing.T) {
	dir := t.TempDir()
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	l.(*joalConfigLoader).clientDownloader = &mockedClientDownloader{}
	if err != nil {
		t.Fatal(err)
	}

	runtimeConfig := RuntimeConfig{}.Default()
	runtimeConfig.Client = "HELLO :D"
	marshalled, err := yaml.Marshal(runtimeConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, runtimeConfigFile), marshalled, 0755); err != nil {
		t.Fatal(err)
	}

	config, err := l.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, runtimeConfig, config.RuntimeConfig)
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldReturnProperJoalConfig(t *testing.T) {
	dir := t.TempDir()
	l, err := NewJoalConfigLoader(dir, &http.Client{})
	l.(*joalConfigLoader).clientDownloader = &mockedClientDownloader{}
	if err != nil {
		t.Fatal(err)
	}

	config, err := l.LoadConfigAndInitIfNeeded()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, RuntimeConfig{}.Default(), config.RuntimeConfig)
	assert.Equal(t, filepath.Join(dir, torrentFolder), config.TorrentsDir)
	assert.Equal(t, filepath.Join(dir, archivedTorrentFolders), config.ArchivedTorrentsDir)
	assert.Equal(t, filepath.Join(dir, clientsFolder), config.ClientsDir)
}
