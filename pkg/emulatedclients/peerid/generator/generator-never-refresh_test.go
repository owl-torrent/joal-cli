package generator

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestUnmarshalNeverRefreshGenerator(t *testing.T) {
	yamlString := `---
type: NEVER_REFRESH
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
	assert.IsType(t, &NeverRefreshGenerator{}, generator.IPeerIdGenerator)
}

func TestGenerateNeverRefresh(t *testing.T) {
	generator := &NeverRefreshGenerator{
		value: nil,
	}
	_ = generator.afterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		generator.get(dumbAlg, infoHash, tracker.None)
	}

	assert.Equal(t, 1, dumbAlg.counter, "Should have been called once")
}

type DumbAlgorithm struct {
	counter int
}

func (d *DumbAlgorithm) Generate() peerid.PeerId {
	d.counter++
	return [20]byte{}
}
func (d *DumbAlgorithm) AfterPropertiesSet() error {
	return nil
}
