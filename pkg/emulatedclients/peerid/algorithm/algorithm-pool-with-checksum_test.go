package algorithm

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestUnmarshalPoolWithChecksumAlgorithm(t *testing.T) {
	yamlString := `---
type: CHAR_POOL_WITH_CHECKSUM
prefix: -TR284Z-
charactersPool: 0123456789abcdefghijklmnopqrstuvwxyz
`
	algorithm := &PeerIdAlgorithm{}
	err := yaml.Unmarshal([]byte(yamlString), algorithm)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	assert.IsType(t, &PoolWithChecksumAlgorithm{}, algorithm.impl)
	assert.Equal(t, algorithm.impl.(*PoolWithChecksumAlgorithm).Prefix, `-TR284Z-`)
	assert.Equal(t, algorithm.impl.(*PoolWithChecksumAlgorithm).CharactersPool, `0123456789abcdefghijklmnopqrstuvwxyz`)
}

func TestGeneratePoolWithChecksumAlgorithm(t *testing.T) {
	alg := &PoolWithChecksumAlgorithm{
		Prefix:         "-TR2820-",
		CharactersPool: "0123456789abcdefghijklmnopqrstuvwxyz",
	}
	_ = alg.AfterPropertiesSet()

	scenarios := []struct {
		randomSource []byte
		expect       string
	}{
		{randomSource: []byte{250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250}, expect: "-TR2820-yyyyyyyyyyym"},
		{randomSource: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, expect: "-TR2820-000000000000"},
		{randomSource: []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}, expect: "-TR2820-333333333333"},
		{randomSource: []byte{128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128}, expect: "-TR2820-kkkkkkkkkkkw"},
		{randomSource: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, expect: "-TR2820-11111111111p"},
		{randomSource: []byte{26, 200, 124, 39, 84, 248, 3, 159, 64, 239, 0}, expect: "-TR2820-qkg3cw3fsn02"},
	}

	for i := 0; i < len(scenarios); i++ {
		alg.RandomSource = bytes.NewReader(scenarios[i].randomSource)
		assert.Equal(t, scenarios[i].expect, fmt.Sprintf("%s", alg.Generate()))
	}
}
