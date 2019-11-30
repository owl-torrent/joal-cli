package emulatedclients

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"testing"
)

func TestEmulatedClient_Unmarshall(t *testing.T) {
	file, _ := os.Open(filepath.Join("testdata", "client.yml"))

	var client EmulatedClient
	err := yaml.NewDecoder(file).Decode(&client)
	if err != nil {
		t.Fatalf("failed to unmarshall EmulatedClient: %v", err)
	}

	err = client.AfterPropertiesSet()
	if err != nil {
		t.Fatalf("failed to validate EmulatedClient: %v", err)
	}

	assert.Equal(t, "qBittorrent", client.Name)
	assert.Equal(t, "3.3.1", client.Version)
	assert.Equal(t, int32(200), client.NumWant)
	assert.Equal(t, int32(0), client.NumWantOnStop)
	assert.NotNil(t, client.KeyGenerator)
	assert.NotNil(t, client.PeerIdGenerator)
	assert.NotNil(t, client.Announcer)
	assert.NotNil(t, client.Listener)
}
