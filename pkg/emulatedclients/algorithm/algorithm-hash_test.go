package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/casing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestUnmarshalHashAlgorithm(t *testing.T) {
	yamlString := `---
type: HASH
trimLeadingZeroes: true
maxLength: 8
case: lower
`
	algorithm := &Algorithm{}
	err := yaml.Unmarshal([]byte(yamlString), algorithm)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = algorithm.AfterPropertiesSet()
	assert.IsType(t, &HashAlgorithm{}, algorithm.impl)
	assert.True(t, algorithm.impl.(*HashAlgorithm).TrimLeadingZeroes)
	assert.Equal(t, 8, algorithm.impl.(*HashAlgorithm).MaxLength)
	assert.Equal(t, casing.Lower, algorithm.impl.(*HashAlgorithm).Case)
}

func TestHashAlgorithm_GenerateShouldGenerateHashes(t *testing.T) {
	algorithm := &HashAlgorithm{
		TrimLeadingZeroes: false,
		MaxLength:         8,
		Case:              casing.None,
	}

	for i := 0; i < 30; i++ {
		res := algorithm.Generate()
		assert.Regexp(t, "^[A-F0-9]{8}$", res)
	}
}

func TestHashAlgorithm_GenerateShouldRespectMaxLength(t *testing.T) {
	algorithm := &HashAlgorithm{
		TrimLeadingZeroes: false,
		MaxLength:         9,
		Case:              casing.None,
	}

	assert.Len(t, algorithm.Generate(), 9)
}

func TestHashAlgorithm_GenerateShouldApplyCase(t *testing.T) {
	assert.Regexp(t, "^[A-F0-9]{8}$", (&HashAlgorithm{
		TrimLeadingZeroes: false,
		MaxLength:         8,
		Case:              casing.Upper,
	}).Generate())
	assert.Regexp(t, "^[a-f0-9]{8}$", (&HashAlgorithm{
		TrimLeadingZeroes: false,
		MaxLength:         8,
		Case:              casing.Lower,
	}).Generate())
}
