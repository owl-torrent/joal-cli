package algorithm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashAlgorithm_GenerateShouldGenerateHashes(t *testing.T) {
	algorithm := &HashAlgorithm{
		TrimLeadingZeroes: false,
		MaxLength:         8,
		Case:              None,
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
		Case:              None,
	}

	assert.Len(t, algorithm.Generate(), 9)
}

func TestHashAlgorithm_GenerateShouldApplyCase(t *testing.T) {
	assert.Regexp(t, "^[A-F0-9]{8}$", (&HashAlgorithm{
		TrimLeadingZeroes: false,
		MaxLength:         8,
		Case:              Upper,
	}).Generate())
	assert.Regexp(t, "^[a-f0-9]{8}$", (&HashAlgorithm{
		TrimLeadingZeroes: false,
		MaxLength:         8,
		Case:              Lower,
	}).Generate())
}
