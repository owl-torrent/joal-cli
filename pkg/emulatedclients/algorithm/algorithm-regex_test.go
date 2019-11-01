package algorithm

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestUnmarshalRegexAlgorithm(t *testing.T) {
	yamlString := `---
type: REGEX
pattern: ^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$
`
	algorithm := &algorithm{}
	err := yaml.Unmarshal([]byte(yamlString), algorithm)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = algorithm.AfterPropertiesSet()
	assert.IsType(t, &RegexPatternAlgorithm{}, algorithm.Impl)
	assert.Equal(t, algorithm.Impl.(*RegexPatternAlgorithm).Pattern, `^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$`)
}

func TestGenerateRegexAlgorithm(t *testing.T) {
	pattern := `^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$`
	alg := &RegexPatternAlgorithm{
		Pattern: pattern,
	}
	_ = alg.AfterPropertiesSet()

	for i := 0; i < 500; i++ {
		assert.Regexp(t, pattern, alg.Generate())
	}
}

func TestGenerateRegexAlgorithmShouldBeRandom(t *testing.T) {
	pattern := `^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$`
	alg := &RegexPatternAlgorithm{
		Pattern: pattern,
	}
	_ = alg.AfterPropertiesSet()

	set := make(map[string]bool)
	for i := 0; i < 500; i++ {
		set[alg.Generate()] = true
	}
	assert.Greater(t, len(set), 300)
}
