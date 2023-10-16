package generator

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestUnmarshalAlwaysRefreshGenerator(t *testing.T) {
	yamlString := `---
type: ALWAYS_REFRESH
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
	assert.IsType(t, &AlwaysRefreshGenerator{}, generator.IPeerIdGenerator)
}

func TestGenerateAlwaysRefresh(t *testing.T) {
	generator := &AlwaysRefreshGenerator{}
	_ = generator.afterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		_ = generator.get(dumbAlg, infoHash, tracker.None)
	}

	assert.Equal(t, 500, dumbAlg.counter, "Should have been called 500 times")
}
