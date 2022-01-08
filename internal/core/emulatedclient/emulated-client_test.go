package emulatedclient

import (
	"github.com/anthonyraymond/joal-cli/internal/core/announcer"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/casing"
	keyAlgorithm "github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key/algorithm"
	keyGenerator "github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key/generator"
	peerIdAlgorithm "github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/peerid/algorithm"
	peerIdGenerator "github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/peerid/generator"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/urlencoder"
	"github.com/anthonyraymond/joal-cli/internal/utils/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
)

func TestEmulatedClient_Unmarshall(t *testing.T) {
	file, _ := os.Open(filepath.Join("testdata", "client.yml"))

	var client EmulatedClient
	err := yaml.NewDecoder(file).Decode(&client)
	if err != nil {
		t.Fatalf("failed to unmarshall EmulatedClient: %v", err)
	}

	err = client.AfterPropertiesSet(nil)
	if err != nil {
		t.Fatalf("failed to validate EmulatedClient: %v", err)
	}

	assert.Equal(t, "qBittorrent", client.Name)
	assert.Equal(t, "3.3.1", client.Version)
	assert.Equal(t, int32(200), client.NumWant)
	assert.Equal(t, int32(0), client.NumWantOnStop)
	assert.NotNil(t, client.KeyGenerator)
	assert.NotNil(t, client.PeerIdGenerator)
	assert.NotNil(t, client.Announcer)
	assert.NotNil(t, client.Listener)
}

func TestEmulatedClient_ShouldValidate(t *testing.T) {
	type args struct {
		Client EmulatedClient
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		failingField     string
		failingTag       string
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithEmptyName", args: args{Client: EmulatedClient{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "EmulatedClient.Name", ErrorTag: "required"}},
		{name: "shouldFailWithEmptyVersion", args: args{Client: EmulatedClient{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "EmulatedClient.Version", ErrorTag: "required"}},
		{name: "shouldFailWithEmptyKeyGenerator", args: args{Client: EmulatedClient{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "EmulatedClient.KeyGenerator", ErrorTag: "required"}},
		{name: "shouldFailWithEmptyPeerIdGenerator", args: args{Client: EmulatedClient{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "EmulatedClient.PeerIdGenerator", ErrorTag: "required"}},
		{name: "shouldFailWithEmptyNumWant", args: args{Client: EmulatedClient{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "EmulatedClient.NumWant", ErrorTag: "min"}},
		{name: "shouldFailWithEmptyAnnouncer", args: args{Client: EmulatedClient{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "EmulatedClient.Announcer", ErrorTag: "required"}},
		{name: "shouldFailWithEmptyListener", args: args{Client: EmulatedClient{}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "EmulatedClient.Listener", ErrorTag: "required"}},
		{name: "shouldValidate", args: args{Client: EmulatedClient{
			Name:    "ok",
			Version: "ok",
			KeyGenerator: &keyGenerator.KeyGenerator{
				IKeyGenerator: &keyGenerator.NeverRefreshGenerator{},
				Algorithm: &keyAlgorithm.KeyAlgorithm{
					IKeyAlgorithm: &keyAlgorithm.NumRangeAsHexAlgorithm{Min: 10, Max: 20},
				},
			},
			PeerIdGenerator: &peerIdGenerator.PeerIdGenerator{
				IPeerIdGenerator: &peerIdGenerator.NeverRefreshGenerator{},
				Algorithm: &peerIdAlgorithm.PeerIdAlgorithm{
					IPeerIdAlgorithm: &peerIdAlgorithm.RegexPatternAlgorithm{Pattern: "[abc]{20}"},
				},
			},
			NumWant:       200,
			NumWantOnStop: 0,
			AnnounceCapabilities: AnnounceCapabilities{
				SupportAnnounceList:         true,
				AnnounceToAllTiers:          true,
				AnnounceToAllTrackersInTier: true,
			},
			Announcer: &announcer.Announcer{
				Http: &announcer.HttpAnnouncer{
					UrlEncoder:     urlencoder.UrlEncoder{EncodedHexCase: casing.Lower},
					Query:          "aaa",
					RequestHeaders: []announcer.HttpRequestHeader{},
				},
				Udp: nil,
			},
			Listener: &Listener{Port: Port{Min: 5050, Max: 8080}},
		}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.New().Struct(tt.args.Client)
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
