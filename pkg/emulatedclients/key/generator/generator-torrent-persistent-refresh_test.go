package generator

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
	"time"
)

func TestUnmarshalTorrentPersistentRefreshGenerator(t *testing.T) {
	yamlString := `---
type: TORRENT_PERSISTENT_REFRESH
algorithm:
  type: NUM_RANGE_ENCODED_AS_HEXADECIMAL
  min: 1
  max: 2
`
	generator := &KeyGenerator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TorrentPersistentGenerator{}, generator.impl)
}

func TestGenerate_TorrentPersistentRefresh_ShouldProvideSingleValuePerTorrent(t *testing.T) {
	generator := &TorrentPersistentGenerator{}
	_ = generator.AfterPropertiesSet()

	infoHashA := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	infoHashB := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	dumbAlg := &DumbAlgorithm{}

	generator.Get(dumbAlg, infoHashA, tracker.None)
	assert.Equal(t, 1, dumbAlg.counter, "Should have been called once for infohashA")
	generator.Get(dumbAlg, infoHashB, tracker.None)
	assert.Equal(t, 2, dumbAlg.counter, "Should have been called once for infohashB")

	for i := 0; i < 500; i++ {
		generator.Get(dumbAlg, infoHashA, tracker.None)
		generator.Get(dumbAlg, infoHashB, tracker.None)
	}

	assert.Equal(t, 2, dumbAlg.counter, "Should not have been called anymore")
	assert.Equal(t, generator.entries[infoHashA].Get(), generator.Get(dumbAlg, infoHashA, tracker.None))
	assert.Equal(t, generator.entries[infoHashB].Get(), generator.Get(dumbAlg, infoHashB, tracker.None))
}

func TestGenerate_TorrentPersistentRefresh_ShouldEvictOldEntries(t *testing.T) {
	generator := &TorrentPersistentGenerator{}
	_ = generator.AfterPropertiesSet()
	generator.evictAfter = 1 * time.Millisecond

	infoHashA := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	infoHashB := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	dumbAlg := &DumbAlgorithm{}

	// Add infohash A (old one)
	generator.entries[infoHashA] = AccessAwareKeyNewSince(1, time.Now().Add(-10*time.Hour))
	// Add infohash B (now)
	generator.entries[infoHashB] = AccessAwareKeyNewSince(2, time.Now())

	for i := 0; i < 120; i++ {
		// work on B to ensure the cleaning counter has revolute as least once
		_ = generator.Get(dumbAlg, infoHashB, tracker.None)
	}

	// pause to let the cleaning goroutine start
	time.Sleep(5 * time.Millisecond)
	// claim a write lock to ensure that the cleaning goroutine is over
	generator.lock.Lock()
	generator.lock.Unlock()

	// ensure A is no longer present in the map
	assert.Len(t, generator.entries, 1)
	assert.NotContains(t, generator.entries, infoHashA)
}

func TestGenerate_TorrentPersistentRefresh_ShouldExpendMapSize(t *testing.T) {
	generator := &TorrentPersistentGenerator{}
	_ = generator.AfterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		_ = generator.Get(dumbAlg, infoHash, tracker.None)
	}

	// ensure A is no longer present in the map
	assert.Len(t, generator.entries, 500)
}
