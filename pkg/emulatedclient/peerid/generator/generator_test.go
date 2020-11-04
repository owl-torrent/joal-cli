package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient/peerid"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient/peerid/algorithm"
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
  type: REGEX
  pattern: ^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$
type: TORRENT_PERSISTENT_REFRESH
`
	generator := &PeerIdGenerator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TorrentPersistentGenerator{}, generator.IPeerIdGenerator)
	assert.NotNil(t, generator.Algorithm)
}

type validAbleKeyGenerator struct {
	Field string `validate:"required"`
}

func (a *validAbleKeyGenerator) get(algorithm.IPeerIdAlgorithm, torrent.InfoHash, tracker.AnnounceEvent) peerid.PeerId {
	return [20]byte{}
}
func (a *validAbleKeyGenerator) afterPropertiesSet() error { return nil }

type validAbleKeyAlg struct {
	Field string `validate:"required"`
}

func (a *validAbleKeyAlg) Generate() peerid.PeerId   { return [20]byte{} }
func (a *validAbleKeyAlg) AfterPropertiesSet() error { return nil }

func TestPeerIdGenerator_ShouldValidate(t *testing.T) {
	type args struct {
		Gen PeerIdGenerator
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithInvalidGenerator", args: args{Gen: PeerIdGenerator{IPeerIdGenerator: &validAbleKeyGenerator{}, Algorithm: &validAbleKeyAlg{Field: "ok"}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "PeerIdGenerator.IPeerIdGenerator.Field", ErrorTag: "required"}},
		{name: "shouldFailWithInvalidAlg", args: args{Gen: PeerIdGenerator{IPeerIdGenerator: &validAbleKeyGenerator{Field: "ok"}, Algorithm: &validAbleKeyAlg{}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "PeerIdGenerator.Algorithm.Field", ErrorTag: "required"}},
		{name: "shouldBeValid", args: args{Gen: PeerIdGenerator{IPeerIdGenerator: &validAbleKeyGenerator{Field: "ok"}, Algorithm: &validAbleKeyAlg{"ok"}}}, wantErr: false},
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
	aas := AccessAwarePeerIdNew([20]byte{0x25})
	aas.lastAccessed = time.Now().Add(-1 * time.Hour) // offset last access

	assert.Equal(t, peerid.PeerId([20]byte{0x25}), aas.Get())
	assert.Less(t, aas.LastAccess().Minutes(), float64(1)) // last access was refreshed and is less than 1m (initial value was 60 min)
}

func TestAccessAwareString_AccessAwareStringNewSince(t *testing.T) {
	expectedTime := time.Now().Add(-80 * time.Minute)
	aas := AccessAwarePeerIdNewSince([20]byte{0xff}, expectedTime)

	assert.Greater(t, aas.LastAccess().Milliseconds(), (79 * time.Minute).Milliseconds()) // last access was refreshed and is less than 1m (initial value was 60 min)
	assert.Equal(t, peerid.PeerId([20]byte{0xff}), aas.Get())
	assert.Less(t, aas.LastAccess().Minutes(), float64(1))
}
