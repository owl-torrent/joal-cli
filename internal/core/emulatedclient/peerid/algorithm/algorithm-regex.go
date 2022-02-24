package algorithm

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/peerid"
	"github.com/lucasjones/reggen"
)

type RegexPatternAlgorithm struct {
	Pattern   string            `yaml:"pattern" validate:"required"`
	generator *reggen.Generator `yaml:"-"`
}

func (r *RegexPatternAlgorithm) Generate() peerid.PeerId {
	pidStr := r.generator.Generate(10)
	var pid peerid.PeerId
	copy(pid[0:peerid.Length], pidStr)
	return pid
}

func (r *RegexPatternAlgorithm) AfterPropertiesSet() error {
	generator, err := reggen.NewGenerator(r.Pattern)
	if err != nil {
		return fmt.Errorf("Bad regex pattern for algorithm generator RegexPatternAlgorithm: %w", err)
	}
	r.generator = generator

	return nil
}
