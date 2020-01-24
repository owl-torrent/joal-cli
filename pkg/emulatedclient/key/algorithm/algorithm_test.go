package algorithm

import (
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient/key"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestKeyAlgorithm_ShouldUnmarshal(t *testing.T) {
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
	assert.IsType(t, &NumRangeAsHexAlgorithm{}, algorithm.IKeyAlgorithm)
	assert.Equal(t, uint32(1), algorithm.IKeyAlgorithm.(*NumRangeAsHexAlgorithm).Min)
	assert.Equal(t, uint32(350), algorithm.IKeyAlgorithm.(*NumRangeAsHexAlgorithm).Max)
}

type validAbleKeyAlg struct {
	Field string `validate:"required"`
}

func (a *validAbleKeyAlg) Generate() key.Key         { return 0 }
func (a *validAbleKeyAlg) AfterPropertiesSet() error { return nil }

func TestKeyAlgorithm_ShouldValidate(t *testing.T) {
	type args struct {
		Alg KeyAlgorithm
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithInvalidNestedField", args: args{Alg: KeyAlgorithm{IKeyAlgorithm: &validAbleKeyAlg{}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "KeyAlgorithm.IKeyAlgorithm.Field", ErrorTag: "required"}},
		{name: "shouldNotFailWithValidNestedField", args: args{Alg: KeyAlgorithm{IKeyAlgorithm: &validAbleKeyAlg{Field: "ok"}}}, wantErr: false},
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
