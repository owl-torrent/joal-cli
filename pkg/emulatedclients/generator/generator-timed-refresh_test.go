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

func TestUnmarshalTimedRefreshGenerator(t *testing.T) {
	yamlString := `---
type: TIMED_REFRESH
refreshEvery: 1ms
algorithm:
  type: REGEX
  pattern: ^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$
`
	generator := &Generator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TimedRefreshGenerator{}, generator.impl)
	assert.Equal(t, 1*time.Millisecond, generator.impl.(*TimedRefreshGenerator).RefreshEvery)
}

func TestGenerate_TimedRefresh_ShouldNotGenerateUntilTimerExpires(t *testing.T) {
	generator := &TimedRefreshGenerator{
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

func TestGenerate_TimedRefresh_ShouldRegenerateWhenTimerExpires(t *testing.T) {
	generator := &TimedRefreshGenerator{
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
