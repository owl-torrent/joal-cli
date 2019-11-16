package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/lucasjones/reggen"
	"github.com/pkg/errors"
)

type RegexPatternAlgorithm struct {
	Pattern   string            `yaml:"pattern"`
	generator *reggen.Generator `yaml:"-"`
}

func (r *RegexPatternAlgorithm) Generate() peerid.PeerId {
	pidStr := r.generator.Generate(10)
	var pid peerid.PeerId
	copy(pid[0:peerid.Length], pidStr)
	return pid
}

func (r *RegexPatternAlgorithm) AfterPropertiesSet() error {
	if len(r.Pattern) == 0 {
		return errors.New("RegexPatternAlgorithm can not have a empty pattern")
	}
	generator, err := reggen.NewGenerator(r.Pattern)
	if err != nil {
		return errors.Wrap(err, "Bad regex pattern for algorithm generator RegexPatternAlgorithm")
	}
	r.generator = generator

	return nil
}
