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
	generator := &generator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TimedRefreshGenerator{}, generator.Impl)
	assert.Equal(t, 1*time.Millisecond, generator.Impl.(*TimedRefreshGenerator).RefreshEvery)
}

func TestGenerate_TimedRefresh_ShouldNotGenerateUntilTimerExpires(t *testing.T) {
	generator := &TimedRefreshGenerator{
		RefreshEvery: 10 * time.Hour,
	}
	_ = generator.AfterPropertiesSet()

	valueSet := make(map[string]bool, 1)
	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		valueSet[generator.Get(dumbAlg, infoHash, tracker.None)] = true
	}

	assert.Equal(t, 1, dumbAlg.counter, "Should have been called once")
	assert.Len(t, valueSet, 1) // has provider unique value
}

func TestGenerate_TimedRefresh_ShouldRegenerateWhenTimerExpires(t *testing.T) {
	generator := &TimedRefreshGenerator{
		RefreshEvery: 1 * time.Millisecond,
	}
	_ = generator.AfterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	generator.Get(dumbAlg, metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), tracker.None)
	time.Sleep(20 * time.Millisecond)
	generator.Get(dumbAlg, metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"), tracker.None)

	assert.Greater(t, dumbAlg.counter, 1, "Should have been called more than once")
}
