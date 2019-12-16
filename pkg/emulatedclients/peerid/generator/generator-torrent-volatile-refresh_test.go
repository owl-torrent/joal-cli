package generator

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
	"time"
)

func TestUnmarshalTorrentVolatileRefreshGenerator(t *testing.T) {
	yamlString := `---
type: TORRENT_VOLATILE_REFRESH
algorithm:
  type: REGEX
  pattern: ^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$
`
	generator := &PeerIdGenerator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TorrentVolatileGenerator{}, generator.IPeerIdGenerator)
}

func TestGenerate_TorrentVolatileRefresh_ShouldProvideSingleValuePerTorrent(t *testing.T) {
	generator := &TorrentVolatileGenerator{}
	_ = generator.afterPropertiesSet()

	infoHashA := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	infoHashB := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	dumbAlg := &DumbAlgorithm{}

	generator.get(dumbAlg, infoHashA, tracker.None)
	assert.Equal(t, 1, dumbAlg.counter, "Should have been called once for infohashA")
	generator.get(dumbAlg, infoHashB, tracker.None)
	assert.Equal(t, 2, dumbAlg.counter, "Should have been called once for infohashB")

	for i := 0; i < 500; i++ {
		generator.get(dumbAlg, infoHashA, tracker.None)
		generator.get(dumbAlg, infoHashB, tracker.None)
	}

	assert.Equal(t, 2, dumbAlg.counter, "Should not have been called anymore")
	assert.Equal(t, generator.entries[infoHashA].Get(), generator.get(dumbAlg, infoHashA, tracker.None))
	assert.Equal(t, generator.entries[infoHashB].Get(), generator.get(dumbAlg, infoHashB, tracker.None))
}

func TestGenerate_TorrentVolatileRefresh_ShouldEvictOldEntries(t *testing.T) {
	generator := &TorrentVolatileGenerator{}
	_ = generator.afterPropertiesSet()
	generator.evictAfter = 3600 * time.Second

	infoHashA := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	infoHashB := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	dumbAlg := &DumbAlgorithm{}

	// Add infohash A (old one)
	generator.entries[infoHashA] = AccessAwarePeerIdNewSince([20]byte{1}, time.Now().Add(-10*time.Hour))
	// Add infohash B (now)
	generator.entries[infoHashB] = AccessAwarePeerIdNewSince([20]byte{2}, time.Now())

	for i := 0; i < 120; i++ {
		// work on B to ensure the cleaning counter has revolute as least once
		_ = generator.get(dumbAlg, infoHashB, tracker.None)
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

func TestGenerate_TorrentVolatileRefresh_ShouldEvictAfterStop(t *testing.T) {
	generator := &TorrentVolatileGenerator{}
	_ = generator.afterPropertiesSet()
	generator.evictAfter = 1 * time.Millisecond

	infoHashA := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	infoHashB := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	dumbAlg := &DumbAlgorithm{}

	// Add infohash A (old one)
	generator.entries[infoHashA] = AccessAwarePeerIdNewSince([20]byte{1}, time.Now())
	// Add infohash B (now)
	generator.entries[infoHashB] = AccessAwarePeerIdNewSince([20]byte{2}, time.Now())

	// returned value on stopped must be the same as previous
	value := generator.get(dumbAlg, infoHashA, tracker.Stopped)
	assert.Equal(t, peerid.PeerId([20]byte{1}), value, "Value must not have been changed on stop")

	// ensure A is no longer present in the map
	assert.Len(t, generator.entries, 1)
	assert.NotContains(t, generator.entries, infoHashA)
}

func TestGenerate_TorrentVolatileRefresh_ShouldExpendMapSize(t *testing.T) {
	generator := &TorrentVolatileGenerator{}
	_ = generator.afterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		_ = generator.get(dumbAlg, infoHash, tracker.None)
	}

	// ensure A is no longer present in the map
	assert.Len(t, generator.entries, 500)
}
