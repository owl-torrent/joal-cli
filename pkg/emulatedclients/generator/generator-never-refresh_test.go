package generator

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
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
	generator := &Generator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &NeverRefreshGenerator{}, generator.impl)
}

func TestGenerateNeverRefresh(t *testing.T) {
	generator := &NeverRefreshGenerator{
		value: nil,
	}
	_ = generator.AfterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		generator.Get(dumbAlg, infoHash, tracker.None)
	}

	assert.Equal(t, 1, dumbAlg.counter, "Should have been called once")
}

type DumbAlgorithm struct {
	counter int
}

func (d *DumbAlgorithm) Generate() string {
	d.counter++
	return "a"
}
func (d *DumbAlgorithm) AfterPropertiesSet() error {
	return nil
}
