package emulatedclient

import (
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListener_Unmarshall(t *testing.T) {
	yamlString := `---
port:
  min: 1
  max: 2
`
	listener := &Listener{}
	err := yaml.Unmarshal([]byte(yamlString), listener)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	assert.Equal(t, uint16(1), listener.Port.Min)
	assert.Equal(t, uint16(2), listener.Port.Max)
}

func TestListener_ShouldValidate(t *testing.T) {
	type args struct {
		Listener Listener
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		errorDescription testutils.ErrorDescription
	}{
		{name: "shouldFailWithMin0", args: args{Listener: Listener{Port: Port{Min: 0, Max: 2}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Listener.Port.Min", ErrorTag: "min"}},
		{name: "shouldFailWithMax0", args: args{Listener: Listener{Port: Port{Min: 2, Max: 0}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Listener.Port.Max", ErrorTag: "min"}},
		{name: "shouldFailMaxLessThanMin", args: args{Listener: Listener{Port: Port{Min: 9800, Max: 500}}}, wantErr: true, errorDescription: testutils.ErrorDescription{ErrorFieldPath: "Listener.Port.Max", ErrorTag: "gtefield"}},
		{name: "shouldNotFailMaxEqualToMin", args: args{Listener: Listener{Port: Port{Min: 9800, Max: 9800}}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.New().Struct(tt.args.Listener)
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

func TestListener_Start(t *testing.T) {
	listener := Listener{
		Port: Port{Min: 8000, Max: 8100},
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("1.1.1.1"))
	}))
	defer s.Close()

	publicIpProviders = []string{s.URL}

	err := listener.Start()
	if err != nil {
		t.Fatalf("failed to get public ip: %v", err)
	}
	assert.Equal(t, net.ParseIP("1.1.1.1").String(), listener.ip.String())
}

func TestListener_getPublicIpShouldFallbackThroughUrl(t *testing.T) {
	listener := Listener{
		Port: Port{Min: 8000, Max: 8100},
	}

	failingServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", 500)
	}))
	defer failingServ.Close()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("1.1.1.1"))
	}))
	defer s.Close()

	publicIpProviders = []string{failingServ.URL, s.URL}

	err := listener.Start()
	if err != nil {
		t.Fatalf("failed to get public ip: %v", err)
	}
	assert.Equal(t, net.ParseIP("1.1.1.1").String(), listener.ip.String())
}
