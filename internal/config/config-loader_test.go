package config

import "testing"

type mockedClientDownloader struct {
	isInstalled func() (bool, error)
	install     func() error
}

func (d *mockedClientDownloader) IsInstalled() (bool, error) {
	if d.isInstalled != nil {
		return d.isInstalled()
	}
	return true, nil
}

func (d *mockedClientDownloader) Install() error {
	if d.install != nil {
		return d.install()
	}
	return nil
}

func TestNewJoalConfigLoader_ShouldCreateNewConfigLoader(t *testing.T) {
	t.Fatal("not implemented")
}

func TestInitialSetup_ShouldCreateAllRequiredAndConfigFileFoldersIfMissing(t *testing.T) {
	t.Fatal("not implemented")
}

func TestHasInitialSetup_ShouldFailIfMissingTorrentFolder(t *testing.T) {
	t.Fatal("not implemented")
}

func TestHasInitialSetup_ShouldFailIfMissingArchiveTorrentFolder(t *testing.T) {
	t.Fatal("not implemented")
}

func TestHasInitialSetup_ShouldFailIfMissingClientFolder(t *testing.T) {
	t.Fatal("not implemented")
}

func TestHasInitialSetup_ShouldFailIfMissingRuntimeConfig(t *testing.T) {
	t.Fatal("not implemented")
}

func TestReadRuntimeConfigOrDefault_ShouldReturnDefaultConfigIfFolderDoesNotExists(t *testing.T) {
	t.Fatal("not implemented")
}

func TestReadRuntimeConfigOrDefault_ShouldReturnDefaultConfigIfFileDoesNotExists(t *testing.T) {
	t.Fatal("not implemented")
}

func TestReadRuntimeConfigOrDefault_ShouldReturnDefaultConfigIfFailsToParse(t *testing.T) {
	t.Fatal("not implemented")
}

func TestReadRuntimeConfigOrDefault_ShouldReturnParseConfigIfFileIsPresent(t *testing.T) {
	t.Fatal("not implemented")
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldCreateFolderStructureIfNotPresent(t *testing.T) {
	t.Fatal("not implemented")
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldAddMissingFolderStructureIfSomeFoldersAreMissing(t *testing.T) {
	t.Fatal("not implemented")
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldDownloadClientsIfNotInstalled(t *testing.T) {
	t.Fatal("not implemented")
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldNotDownloadClientsIfAlreadyInstalled(t *testing.T) {
	t.Fatal("not implemented")
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldReturnDefaultConfigIfRuntimeConfigFileIsNotPresent(t *testing.T) {
	t.Fatal("not implemented")
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldReturnParsedRuntimeConfigIfFileIsPresent(t *testing.T) {
	t.Fatal("not implemented")
}

func TestJoalConfigLoader_LoadConfigAndInitIfNeeded_ShouldReturnProperJoalConfig(t *testing.T) {
	t.Fatal("not implemented")
}
