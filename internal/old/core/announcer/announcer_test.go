package announcer

import (
	"context"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/casing"
	"github.com/anthonyraymond/joal-cli/internal/old/utils/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"testing"
)

func TestAnnouncer_ShouldUnmarshal(t *testing.T) {
	yamlString := `---
http:
  query: dd
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

func TestAnnouncer_ShouldUnmarshalAndBeOverrideable(t *testing.T) {
	yamlString := `---
http:
  urlEncoder:
    encodedHexCase: upper
  query: dd
`
	// TODO: add UDP
	announcer := &Announcer{}
	err := yaml.Unmarshal([]byte(yamlString), announcer)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	assert.NotNil(t, announcer.Http)
	assert.IsType(t, &HttpAnnouncer{}, announcer.Http)
	assert.Equal(t, announcer.Http.(*HttpAnnouncer).Query, "dd")
	assert.Equal(t, announcer.Http.(*HttpAnnouncer).UrlEncoder.EncodedHexCase, casing.Upper)
	/* TODO: Add UDP
	assert.NotNil(t, announcer.Udp)
	assert.IsType(t, &UdpAnnouncer{}, announcer.Udp)
	*/

	yamlString = `---
http:
  query: changingTheQueryShouldNotResetOtherFields
`
	// unmarshall in the same struct to override some fields but not all
	err = yaml.Unmarshal([]byte(yamlString), announcer)
	if err != nil {
		t.Fatalf("Failed to unmarshall override: %+v", err)
	}
	assert.Equal(t, announcer.Http.(*HttpAnnouncer).Query, "changingTheQueryShouldNotResetOtherFields")
	assert.Equal(t, announcer.Http.(*HttpAnnouncer).UrlEncoder.EncodedHexCase, casing.Upper) // should not have changed

	//TODO: Add UDP
}

type validAbleHttpAnnouncer struct {
	Field string `validate:"required"`
}

func (a *validAbleHttpAnnouncer) AfterPropertiesSet(proxyFunc func(*http.Request) (*url.URL, error)) error {
	return nil
}
func (a *validAbleHttpAnnouncer) Announce(url.URL, AnnounceRequest, context.Context) (AnnounceResponse, error) {
	return AnnounceResponse{}, nil
}

type validAbleUdpAnnouncer struct {
	Field string `validate:"required"`
}

func (a *validAbleUdpAnnouncer) AfterPropertiesSet() error { return nil }
func (a *validAbleUdpAnnouncer) Announce(url.URL, AnnounceRequest, context.Context) (AnnounceResponse, error) {
	return AnnounceResponse{}, nil
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
		{name: "shouldFailWithInvalidNestedFieldHttp", args: args{Announcer: Announcer{Http: &validAbleHttpAnnouncer{}, Udp: &validAbleUdpAnnouncer{Field: "ok"}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Announcer.Http.Field", ErrorTag: "required"}},
		{name: "shouldFailWithInvalidNestedFieldUdp", args: args{Announcer: Announcer{Http: &validAbleHttpAnnouncer{"ok"}, Udp: &validAbleUdpAnnouncer{}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Announcer.Udp.Field", ErrorTag: "required"}},
		{name: "shouldFailWithNoUdpOrHttp", args: args{Announcer: Announcer{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Announcer.Http", ErrorTag: "required_without_all"}},
		{name: "shouldNotFailWithOnlyHttp", args: args{Announcer: Announcer{Http: &validAbleHttpAnnouncer{Field: "ok"}}}, wantErr: false},
		{name: "shouldNotFailWithOnlyUdp", args: args{Announcer: Announcer{Udp: &validAbleUdpAnnouncer{Field: "ok"}}}, wantErr: false},
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

type fakeSubHttpAnnouncer struct {
	announce func(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error)
}

func (f *fakeSubHttpAnnouncer) Announce(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error) {
	if f.announce != nil {
		return f.announce(u, announceRequest, ctx)
	}
	return AnnounceResponse{}, nil
}

func (f *fakeSubHttpAnnouncer) AfterPropertiesSet(proxyFunc func(*http.Request) (*url.URL, error)) error {
	return nil
}

type fakeSubUdpAnnouncer struct {
	announce func(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error)
}

func (f *fakeSubUdpAnnouncer) Announce(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error) {
	if f.announce != nil {
		return f.announce(u, announceRequest, ctx)
	}
	return AnnounceResponse{}, nil
}

func (f *fakeSubUdpAnnouncer) AfterPropertiesSet() error {
	return nil
}

func TestAnnouncer_SelectAnnouncerBasedOnUrlScheme(t *testing.T) {
	announceDone := 0

	announcer := &Announcer{
		Http: &fakeSubHttpAnnouncer{announce: func(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error) {
			announceDone++
			if u.Scheme != "http" && u.Scheme != "https" {
				t.Fatal("non http scheme url passed to http announcer")
			}
			return AnnounceResponse{}, nil
		}},
		Udp: &fakeSubUdpAnnouncer{announce: func(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error) {
			announceDone++
			if u.Scheme != "udp" && u.Scheme != "udp4" && u.Scheme != "udp6" {
				t.Fatal("non udp scheme url passed to udp announcer")
			}
			return AnnounceResponse{}, nil
		}},
	}

	_, _ = announcer.Announce(*testutils.MustParseUrl("http://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("https://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("udp://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("udp4://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("udp6://localhost.fr"), AnnounceRequest{}, context.Background())

	assert.Equal(t, 5, announceDone)
}

func TestAnnouncer_AnnounceHttpWithNilHttpAnnouncer(t *testing.T) {
	announcer := &Announcer{
		Udp: &fakeSubUdpAnnouncer{},
	}

	_, err := announcer.Announce(*testutils.MustParseUrl("http://localhost.fr"), AnnounceRequest{}, context.Background())
	if err == nil {
		t.Fatal("should have returned an error")
	}

	assert.Contains(t, err.Error(), "'http' is not supported")
}

func TestAnnouncer_AnnounceUdpWithNilUdpAnnouncer(t *testing.T) {
	announcer := &Announcer{
		Http: &fakeSubHttpAnnouncer{},
	}

	_, err := announcer.Announce(*testutils.MustParseUrl("udp://localhost.fr"), AnnounceRequest{}, context.Background())
	if err == nil {
		t.Fatal("should have returned an error")
	}

	assert.Contains(t, err.Error(), "'udp' is not supported")
}

func TestAnnouncer_AnnounceUnknownScheme(t *testing.T) {
	announcer := &Announcer{
		Http: &fakeSubHttpAnnouncer{},
		Udp:  &fakeSubUdpAnnouncer{},
	}

	_, err := announcer.Announce(*testutils.MustParseUrl("belozic://localhost.fr"), AnnounceRequest{}, context.Background())
	if err == nil {
		t.Fatal("should have returned an error")
	}

	assert.Contains(t, err.Error(), "'belozic' is not supported")
}
