package algorithm

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/peerid"
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestUnmarshalRegexAlgorithm(t *testing.T) {
	yamlString := `---
type: REGEX
pattern: ^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$
`
	algorithm := &PeerIdAlgorithm{}
	err := yaml.Unmarshal([]byte(yamlString), algorithm)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = algorithm.AfterPropertiesSet()
	assert.IsType(t, &RegexPatternAlgorithm{}, algorithm.IPeerIdAlgorithm)
	assert.Equal(t, algorithm.IPeerIdAlgorithm.(*RegexPatternAlgorithm).Pattern, `^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$`)
}

func TestRegexPatternAlgorithm_ShouldValidate(t *testing.T) {
	type args struct {
		Alg RegexPatternAlgorithm
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithEmptyPattern", args: args{Alg: RegexPatternAlgorithm{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "RegexPatternAlgorithm.Pattern", ErrorTag: "required"}},
		{name: "shouldValidate", args: args{Alg: RegexPatternAlgorithm{Pattern: "ok"}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.New().Struct(tt.args.Alg)
			if tt.wantErr == true && err == nil {
				t.Fatal("validation failed, wantErr=true but err is nil")
			}
			if tt.wantErr == false && err != nil {
				t.Fatalf("validation failed, wantErr=false but err is : %v", err)
			}
			if tt.wantErr {
				testutils.AssertValidateError(t, err.(validator.ValidationErrors), tt.errorDescription)
			}
		})
	}
}

func TestGenerateRegexAlgorithm(t *testing.T) {
	pattern := `^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$`
	alg := &RegexPatternAlgorithm{
		Pattern: pattern,
	}
	_ = alg.AfterPropertiesSet()

	for i := 0; i < 500; i++ {
		assert.Regexp(t, pattern, fmt.Sprintf("%s", alg.Generate()))
	}
}

func TestGenerateRegexAlgorithmShouldBeRandom(t *testing.T) {
	pattern := `^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$`
	alg := &RegexPatternAlgorithm{
		Pattern: pattern,
	}
	_ = alg.AfterPropertiesSet()

	set := make(map[peerid.PeerId]bool)
	for i := 0; i < 500; i++ {
		set[alg.Generate()] = true
	}
	assert.Greater(t, len(set), 300)
}
