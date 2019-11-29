package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/anthonyraymond/joal-cli/pkg/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestKeyAlgorithm_ShouldUnmarshal(t *testing.T) {
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
	assert.Equal(t, `^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$`, algorithm.IPeerIdAlgorithm.(*RegexPatternAlgorithm).Pattern)
}

type validAbleKeyAlg struct {
	Field string `validate:"required"`
}

func (a *validAbleKeyAlg) Generate() peerid.PeerId   { return [20]byte{} }
func (a *validAbleKeyAlg) AfterPropertiesSet() error { return nil }

func TestKeyAlgorithm_ShouldValidate(t *testing.T) {
	type args struct {
		Alg PeerIdAlgorithm
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithInvalidNestedField", args: args{Alg: PeerIdAlgorithm{IPeerIdAlgorithm: &validAbleKeyAlg{}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "PeerIdAlgorithm.IPeerIdAlgorithm.Field", ErrorTag: "required"}},
		{name: "shouldNotFailWithValidNestedField", args: args{Alg: PeerIdAlgorithm{IPeerIdAlgorithm: &validAbleKeyAlg{Field: "ok"}}}, wantErr: false},
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
