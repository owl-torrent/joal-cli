package generator

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
	"time"
)

func TestUnmarshalTimedOrAfterStartedAnnounceRefreshGenerator(t *testing.T) {
	yamlString := `---
type: TIMED_OR_AFTER_STARTED_ANNOUNCE_REFRESH
refreshEvery: 1ms
algorithm:
  type: REGEX
  pattern: ^-qB3310-[A-Za-z0-9_~\(\)\!\.\*-]{12}$
`
	generator := &PeerIdGenerator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TimedOrAfterStartedAnnounceRefreshGenerator{}, generator.IPeerIdGenerator)
	assert.Equal(t, 1*time.Millisecond, generator.IPeerIdGenerator.(*TimedOrAfterStartedAnnounceRefreshGenerator).RefreshEvery)
}

func TestTimedOrAfterStartedAnnounceRefresh_ShouldValidate(t *testing.T) {
	type args struct {
		Gen TimedOrAfterStartedAnnounceRefreshGenerator
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithNoDuration", args: args{Gen: TimedOrAfterStartedAnnounceRefreshGenerator{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedOrAfterStartedAnnounceRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldFailWith0nsDuration", args: args{Gen: TimedOrAfterStartedAnnounceRefreshGenerator{RefreshEvery: 0 * time.Nanosecond}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedOrAfterStartedAnnounceRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldFailWith0msDuration", args: args{Gen: TimedOrAfterStartedAnnounceRefreshGenerator{RefreshEvery: 0 * time.Millisecond}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedOrAfterStartedAnnounceRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldFailWith0sDuration", args: args{Gen: TimedOrAfterStartedAnnounceRefreshGenerator{RefreshEvery: 0 * time.Second}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedOrAfterStartedAnnounceRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldValidate", args: args{Gen: TimedOrAfterStartedAnnounceRefreshGenerator{RefreshEvery: 30 * time.Minute}}, wantErr: false},
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

func TestGenerate_TimedOrAfterStartedAnnounceRefresh_ShouldNotGenerateUntilTimerExpires(t *testing.T) {
	generator := &TimedOrAfterStartedAnnounceRefreshGenerator{
		RefreshEvery: 10 * time.Hour,
	}
	_ = generator.afterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	for i := 0; i < 500; i++ {
		infoHash := metainfo.NewHashFromHex(fmt.Sprintf("%dAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", i)[0:40])
		generator.get(dumbAlg, infoHash, tracker.None)
	}

	assert.Equal(t, 1, dumbAlg.counter, "Should have been called once")
}

func TestGenerate_TimedOrAfterStartedAnnounceRefresh_ShouldRegenerateWhenTimerExpires(t *testing.T) {
	generator := &TimedOrAfterStartedAnnounceRefreshGenerator{
		RefreshEvery: 1 * time.Millisecond,
	}
	_ = generator.afterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	infoHash := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	generator.get(dumbAlg, infoHash, tracker.None)
	generator.nextGeneration = time.Now().Add(-1 * time.Second)
	generator.get(dumbAlg, infoHash, tracker.None)

	assert.Greater(t, dumbAlg.counter, 1, "Should have been called more than once")
}

func TestGenerate_TimedOrAfterStartedAnnounceRefresh_ShouldRegenerateWhenAnnounceIsStarted(t *testing.T) {
	generator := &TimedOrAfterStartedAnnounceRefreshGenerator{
		RefreshEvery: 1 * time.Hour,
	}
	_ = generator.afterPropertiesSet()

	dumbAlg := &DumbAlgorithm{}
	infoHash := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	generator.get(dumbAlg, infoHash, tracker.Started)
	generator.get(dumbAlg, infoHash, tracker.Started)
	generator.get(dumbAlg, infoHash, tracker.Started)

	assert.Equal(t, 3, dumbAlg.counter, "Should have been called more than once")
}
