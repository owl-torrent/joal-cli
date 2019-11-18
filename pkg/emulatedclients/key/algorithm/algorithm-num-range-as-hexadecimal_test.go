package algorithm

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestUnmarshalNumRangeAsHexadecimalAlgorithm(t *testing.T) {
	yamlString := `---
type: NUM_RANGE_ENCODED_AS_HEXADECIMAL
min: 1
max: 350
`
	algorithm := &KeyAlgorithm{}
	err := yaml.Unmarshal([]byte(yamlString), algorithm)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = algorithm.AfterPropertiesSet()
	assert.IsType(t, &NumRangeAsHexAlgorithm{}, algorithm.impl)
	assert.Equal(t, uint32(1), algorithm.impl.(*NumRangeAsHexAlgorithm).Min)
	assert.Equal(t, uint32(350), algorithm.impl.(*NumRangeAsHexAlgorithm).Max)
}
