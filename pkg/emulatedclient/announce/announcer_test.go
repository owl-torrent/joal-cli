package announce

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"net/url"
	"strings"
	"testing"
)

func TestAnnouncer_ShouldUnmarshal(t *testing.T) {
	yamlString := `---
announcer:
  http:
`
	// TODO: add UDP
	announcer := &Announcer{}
	err := yaml.Unmarshal([]byte(yamlString), announcer)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	assert.NotNil(t, announcer.Http)
	assert.IsType(t, &HttpAnnouncer{}, announcer.Http)
	/* TODO: Add UDP
	assert.NotNil(t, announcer.Udp)
	assert.IsType(t, &UdpAnnouncer{}, announcer.Udp)
	*/
}

type validAbleAnnouncer struct {
	Field string `validate:"required"`
}

func (a *validAbleAnnouncer) AfterPropertiesSet() error { return nil }
func (a *validAbleAnnouncer) Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error) {
	return tracker.AnnounceResponse{}, nil
}

func TestAnnouncer_ShouldValidate(t *testing.T) {
	type args struct {
		Announcer Announcer
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithInvalidNestedFieldHttp", args: args{Announcer: Announcer{Http: &validAbleAnnouncer{}, Udp: &validAbleAnnouncer{Field: "ok"}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Announcer.Http.Field", ErrorTag: "required"}},
		{name: "shouldFailWithInvalidNestedFieldUdp", args: args{Announcer: Announcer{Http: &validAbleAnnouncer{"ok"}, Udp: &validAbleAnnouncer{}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Announcer.Udp.Field", ErrorTag: "required"}},
		{name: "shouldFailWithNoUdpOrHttp", args: args{Announcer: Announcer{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Announcer.Http", ErrorTag: "required_without_all"}},
		{name: "shouldNotFailWithOnlyHttp", args: args{Announcer: Announcer{Http: &validAbleAnnouncer{Field: "ok"}}}, wantErr: false},
		{name: "shouldNotFailWithOnlyUdp", args: args{Announcer: Announcer{Udp: &validAbleAnnouncer{Field: "ok"}}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.New().Struct(tt.args.Announcer)
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

func TestAnnouncer_AnnounceShouldCallAnnouncerCorrespondingToScheme(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"http://localhost.fr"}}, AnnounceRequest{}, context.Background())
	assert.Equal(t, 1, announcer.Http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"https://localhost.fr"}}, AnnounceRequest{}, context.Background())
	assert.Equal(t, 2, announcer.Http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"udp://localhost.fr"}}, AnnounceRequest{}, context.Background())
	assert.Equal(t, 1, announcer.Udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"udp4://localhost.fr"}}, AnnounceRequest{}, context.Background())
	assert.Equal(t, 2, announcer.Udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"udp6://localhost.fr"}}, AnnounceRequest{}, context.Background())
	assert.Equal(t, 3, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_Announce_ShouldNotDemoteIfSucceed(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	urls := metainfo.AnnounceList{{"http://localhost.fr", "udp://localhost.fr"}}
	expected := metainfo.AnnounceList{{"http://localhost.fr", "udp://localhost.fr"}}
	_, _ = announcer.Announce(&urls, AnnounceRequest{}, context.Background())
	assert.Equal(t, 1, announcer.Http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, expected, urls)
	assert.Equal(t, 0, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_Announce_ShouldPromoteTierAndUrlInTierIfSucceed(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	urls := metainfo.AnnounceList{
		{"http://localhost.fr/fail", "http://localhost.fr/x/fail", "http://localhost.fr/y/fail"},
		{"http://localhost.fr/t2/fail", "http://localhost.fr/t2/x/fail", "http://localhost.fr/t2/y"},
	}
	expected := metainfo.AnnounceList{
		{"http://localhost.fr/t2/y", "http://localhost.fr/t2/fail", "http://localhost.fr/t2/x/fail"},
		{"http://localhost.fr/fail", "http://localhost.fr/x/fail", "http://localhost.fr/y/fail"},
	}
	_, _ = announcer.Announce(&urls, AnnounceRequest{}, context.Background())
	assert.Equal(t, 6, announcer.Http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, expected, urls)
	assert.Equal(t, 0, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

type DumbHttpAnnouncer struct {
	counter int
}

func (a *DumbHttpAnnouncer) AfterPropertiesSet() error { return nil }
func (a *DumbHttpAnnouncer) Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error) {
	a.counter++
	if strings.Contains(url.String(), "fail") {
		return tracker.AnnounceResponse{}, errors.New("asked to fail because url contains 'fail'")
	}
	return tracker.AnnounceResponse{}, nil
}

type DumbUdpAnnouncer struct {
	counter int
}

func (a *DumbUdpAnnouncer) AfterPropertiesSet() error { return nil }
func (a *DumbUdpAnnouncer) Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error) {
	a.counter++
	if strings.Contains(url.String(), "fail") {
		return tracker.AnnounceResponse{}, errors.New("asked to fail because url contains 'fail'")
	}
	return tracker.AnnounceResponse{}, nil
}
