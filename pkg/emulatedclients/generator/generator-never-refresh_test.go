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
	generator := &generator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &NeverRefreshGenerator{}, generator.Impl)
}

func TestGenerateNeverRefresh(t *testing.T) {
	generator := &NeverRefreshGenerator{
		value: nil,
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
