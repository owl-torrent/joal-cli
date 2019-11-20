package emulatedclients

import (
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("1.1.1.1"))
	}))
	defer s.Close()

	publicIpProviders = []string{"http://localhost:15915/noop", failingServ.URL, s.URL}

	err := listener.Start()
	if err != nil {
		t.Fatalf("failed to get public ip: %v", err)
	}
	assert.Equal(t, net.ParseIP("1.1.1.1").String(), listener.ip.String())
}
