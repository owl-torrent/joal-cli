package announce

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"net/url"
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
func (a *validAbleAnnouncer) Announce(url.URL, AnnounceRequest, context.Context) (tracker.AnnounceResponse, error) {
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpAnn := NewMockIHttpAnnouncer(ctrl)
	udpAnn := NewMockIUdpAnnouncer(ctrl)
	announcer := Announcer{
		Http: httpAnn,
		Udp:  udpAnn,
	}

	gomock.InOrder(
		httpAnn.EXPECT().Announce(gomock.Eq(*testutils.MustParseUrl("http://localhost.fr")), gomock.Any(), gomock.Any()).Times(1),
		httpAnn.EXPECT().Announce(gomock.Eq(*testutils.MustParseUrl("https://localhost.fr")), gomock.Any(), gomock.Any()).Times(1),
		udpAnn.EXPECT().Announce(gomock.Eq(*testutils.MustParseUrl("udp://localhost.fr")), gomock.Any(), gomock.Any()).Times(1),
		udpAnn.EXPECT().Announce(gomock.Eq(*testutils.MustParseUrl("udp4://localhost.fr")), gomock.Any(), gomock.Any()).Times(1),
		udpAnn.EXPECT().Announce(gomock.Eq(*testutils.MustParseUrl("udp6://localhost.fr")), gomock.Any(), gomock.Any()).Times(1),
	)

	_, _ = announcer.Announce(*testutils.MustParseUrl("http://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("https://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("udp://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("udp4://localhost.fr"), AnnounceRequest{}, context.Background())
	_, _ = announcer.Announce(*testutils.MustParseUrl("udp6://localhost.fr"), AnnounceRequest{}, context.Background())
}
