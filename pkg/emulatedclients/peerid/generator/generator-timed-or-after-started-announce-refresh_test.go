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

func TestUnmarshalTimedOrAfterStartedAnnounceRefreshGenerator(t *testing.T) {
	yamlString := `---
type: TIMED_OR_AFTER_STARTED_ANNOUNCE_REFRESH
refreshEvery: 1ms
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
	assert.IsType(t, &TimedOrAfterStartedAnnounceRefreshGenerator{}, generator.impl)
	assert.Equal(t, 1*time.Millisecond, generator.impl.(*TimedOrAfterStartedAnnounceRefreshGenerator).RefreshEvery)
}

func TestGenerate_TimedOrAfterStartedAnnounceRefresh_ShouldNotGenerateUntilTimerExpires(t *testing.T) {
	generator := &TimedOrAfterStartedAnnounceRefreshGenerator{
		RefreshEvery: 10 * time.Hour,
	}
	_ = generator.AfterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		generator.Get(dumbAlg, infoHash, tracker.None)
	}

	assert.Equal(t, 1, dumbAlg.counter, "Should have been called once")
}

func TestGenerate_TimedOrAfterStartedAnnounceRefresh_ShouldRegenerateWhenTimerExpires(t *testing.T) {
	generator := &TimedOrAfterStartedAnnounceRefreshGenerator{
		RefreshEvery: 1 * time.Millisecond,
	}
	_ = generator.AfterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	infoHash := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	generator.Get(dumbAlg, infoHash, tracker.None)
	generator.nextGeneration = time.Now().Add(-1 * time.Second)
	generator.Get(dumbAlg, infoHash, tracker.None)

	assert.Greater(t, dumbAlg.counter, 1, "Should have been called more than once")
}

func TestGenerate_TimedOrAfterStartedAnnounceRefresh_ShouldRegenerateWhenAnnounceIsStarted(t *testing.T) {
	generator := &TimedOrAfterStartedAnnounceRefreshGenerator{
		RefreshEvery: 1 * time.Hour,
	}
	_ = generator.AfterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	infoHash := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	generator.Get(dumbAlg, infoHash, tracker.Started)
	generator.Get(dumbAlg, infoHash, tracker.Started)
	generator.Get(dumbAlg, infoHash, tracker.Started)

	assert.Equal(t, 3, dumbAlg.counter, "Should have been called more than once")
}
