package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestNumRangeAsHexAlgorithm_ShouldUnmarshal(t *testing.T) {
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

func TestHttpAnnouncer_ShouldValidate(t *testing.T) {
	type args struct {
		Alg NumRangeAsHexAlgorithm
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithMaxEqual0", args: args{Alg: NumRangeAsHexAlgorithm{Min: 0, Max: 0}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "NumRangeAsHexAlgorithm.Max", ErrorTag: "min"}},
		{name: "shouldFailWithMinGreaterThanMax", args: args{Alg: NumRangeAsHexAlgorithm{Min: 10, Max: 5}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "NumRangeAsHexAlgorithm.Max", ErrorTag: "gtefield"}},
		{name: "shouldNotFailWithMaxEqualMin", args: args{Alg: NumRangeAsHexAlgorithm{Min: 5, Max: 5}}, wantErr: false},
		{name: "shouldNotFailWithMinGreaterThanMax", args: args{Alg: NumRangeAsHexAlgorithm{Min: 5, Max: 10}}, wantErr: false},
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

func TestNumRangeAsHexAlgorithm_Generate(t *testing.T) {
	alg := NumRangeAsHexAlgorithm{
		Min: 1,
		Max: 2,
	}
	err := alg.AfterPropertiesSet()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 500; i++ {
		keyVal := uint32(alg.Generate())
		assert.GreaterOrEqual(t, keyVal, uint32(1))
		assert.LessOrEqual(t, keyVal, uint32(2))
	}
}
