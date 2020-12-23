package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/key"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/key/algorithm"
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
	"time"
)

func TestKeyGenerator_ShouldUnmarshal(t *testing.T) {
	yamlString := `---
algorithm:
  type: NUM_RANGE_ENCODED_AS_HEXADECIMAL
  min: 0
  max: 4294967295
type: TORRENT_PERSISTENT_REFRESH
`
	generator := &KeyGenerator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TorrentPersistentGenerator{}, generator.IKeyGenerator)
	assert.NotNil(t, generator.Algorithm)
}

type validAbleKeyGenerator struct {
	Field string `validate:"required"`
}

func (a *validAbleKeyGenerator) get(algorithm.IKeyAlgorithm, torrent.InfoHash, tracker.AnnounceEvent) key.Key {
	return 0
}
func (a *validAbleKeyGenerator) afterPropertiesSet() error { return nil }

type validAbleKeyAlg struct {
	Field string `validate:"required"`
}

func (a *validAbleKeyAlg) Generate() key.Key         { return 0 }
func (a *validAbleKeyAlg) AfterPropertiesSet() error { return nil }

func TestKeyGenerator_ShouldValidate(t *testing.T) {
	type args struct {
		Gen KeyGenerator
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithInvalidGenerator", args: args{Gen: KeyGenerator{IKeyGenerator: &validAbleKeyGenerator{}, Algorithm: &validAbleKeyAlg{Field: "ok"}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "KeyGenerator.IKeyGenerator.Field", ErrorTag: "required"}},
		{name: "shouldFailWithInvalidAlg", args: args{Gen: KeyGenerator{IKeyGenerator: &validAbleKeyGenerator{Field: "ok"}, Algorithm: &validAbleKeyAlg{}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "KeyGenerator.Algorithm.Field", ErrorTag: "required"}},
		{name: "shouldBeValid", args: args{Gen: KeyGenerator{IKeyGenerator: &validAbleKeyGenerator{Field: "ok"}, Algorithm: &validAbleKeyAlg{"ok"}}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.New().Struct(tt.args.Gen)
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

func TestAccessAwareString_GetShouldRefreshLastAccess(t *testing.T) {
	aas := AccessAwareKeyNew(12)
	aas.lastAccessed = time.Now().Add(-1 * time.Hour) // offset last access

	assert.Equal(t, key.Key(12), aas.Get())
	assert.Less(t, time.Since(aas.lastAccessed).Minutes(), float64(1)) // last access was refreshed and is less than 1m (initial value was 60 min)
}
