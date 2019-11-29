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

func TestUnmarshalTimedRefreshGenerator(t *testing.T) {
	yamlString := `---
type: TIMED_REFRESH
refreshEvery: 1ms
algorithm:
  type: NUM_RANGE_ENCODED_AS_HEXADECIMAL
  min: 1
  max: 2
`
	generator := &KeyGenerator{}
	err := yaml.Unmarshal([]byte(yamlString), generator)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	_ = generator.AfterPropertiesSet()
	assert.IsType(t, &TimedRefreshGenerator{}, generator.IKeyGenerator)
	assert.Equal(t, 1*time.Millisecond, generator.IKeyGenerator.(*TimedRefreshGenerator).RefreshEvery)
}

func TestTimedRefresh_ShouldValidate(t *testing.T) {
	type args struct {
		Gen TimedRefreshGenerator
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithNoDuration", args: args{Gen: TimedRefreshGenerator{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldFailWith0nsDuration", args: args{Gen: TimedRefreshGenerator{RefreshEvery: 0 * time.Nanosecond}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldFailWith0msDuration", args: args{Gen: TimedRefreshGenerator{RefreshEvery: 0 * time.Millisecond}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldFailWith0sDuration", args: args{Gen: TimedRefreshGenerator{RefreshEvery: 0 * time.Second}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "TimedRefreshGenerator.RefreshEvery", ErrorTag: "required"}},
		{name: "shouldValidate", args: args{Gen: TimedRefreshGenerator{RefreshEvery: 30 * time.Minute}}, wantErr: false},
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

func TestGenerate_TimedRefresh_ShouldNotGenerateUntilTimerExpires(t *testing.T) {
	generator := &TimedRefreshGenerator{
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

func TestGenerate_TimedRefresh_ShouldRegenerateWhenTimerExpires(t *testing.T) {
	generator := &TimedRefreshGenerator{
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
