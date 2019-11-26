package emulatedclients

import (
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
}