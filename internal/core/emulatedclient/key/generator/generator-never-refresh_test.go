package generator

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestUnmarshalNeverRefreshGenerator(t *testing.T) {
	yamlString := `---
type: NEVER_REFRESH
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
	_ = generator.afterPropertiesSet()
	assert.IsType(t, &NeverRefreshGenerator{}, generator.IKeyGenerator)
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

func (d *DumbAlgorithm) Generate() key.Key {
	d.counter++
	return 12
}
func (d *DumbAlgorithm) AfterPropertiesSet() error {
	return nil
}
