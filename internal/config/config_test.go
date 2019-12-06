package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigManager_Load(t *testing.T) {
	configFilePath, err := filepath.Abs("./testdata/seed-config.json")
	if err != nil {
		t.Fatal(err)
	}
	manager, err := ConfigManagerNew(configFilePath)
	if err != nil {
		t.Fatal(err)
	}

	conf, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}

	assert.NotZero(t, conf.MinUploadRate)
	assert.NotZero(t, conf.MaxUploadRate)
	assert.NotEmpty(t, conf.Client)
	assert.False(t, conf.RemoveTorrentWithZeroPeers)

	confFromGet, _ := manager.Get()
	assert.Equal(t, confFromGet, conf)
}

func TestConfigManager_Get_ShouldLoadIfNotLoadedYet(t *testing.T) {
	configFilePath, err := filepath.Abs("./testdata/seed-config.json")
	if err != nil {
		t.Fatal(err)
	}
	manager, err := ConfigManagerNew(configFilePath)
	if err != nil {
		t.Fatal(err)
	}

	conf, err := manager.Get()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotZero(t, conf.MinUploadRate)
	assert.NotZero(t, conf.MaxUploadRate)
	assert.NotEmpty(t, conf.Client)
	assert.False(t, conf.RemoveTorrentWithZeroPeers)
}

func TestConfigManager_Save(t *testing.T) {
	configFilePath, err := filepath.Abs("./testdata/tmp-config.json")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.Remove(configFilePath)
	}()

	manager, err := ConfigManagerNew(configFilePath)
	if err != nil {
		t.Fatal(err)
	}

	err = manager.Save(SeedConfig{
		MinUploadRate:              10,
		MaxUploadRate:              10,
		Client:                     "a",
		RemoveTorrentWithZeroPeers: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		t.Fatalf("config file was supposed to exists but it does not: '%s': %v", configFilePath, err)
	}
}
