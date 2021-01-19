package web

import (
	"math"
	"time"
)

type webConfig struct {
	Http      *httpConfig      `yaml:"http"`
	WebSocket *webSocketConfig `yaml:"webSocket"`
	Stomp     *stompConfig     `yaml:"stomp"`
}

// Return a new webConfig with the default values filled in
func (c webConfig) Default() *webConfig {
	return &webConfig{
		Http:      httpConfig{}.Default(),
		WebSocket: webSocketConfig{}.Default(),
		Stomp:     stompConfig{}.Default(),
	}
}

type httpConfig struct {
	Port                     int           `yaml:"port"`
	ReadTimeout              time.Duration `yaml:"readTimeout"`
	ReadHeaderTimeout        time.Duration `yaml:"readHeaderTimeout"`
	WriteTimeout             time.Duration `yaml:"writeTimeout"`
	IdleTimeout              time.Duration `yaml:"idleTimeout"`
	MaxHeaderBytes           int           `yaml:"maxHeaderBytes"`
	WebUiUrl                 string        `yaml:"webUiUrl"`
	HttpApiUrl               string        `yaml:"httpApiUrl"`
	WsNegotiationEndpointUrl string        `yaml:"wsNegotiationEndpointUrl"`
}

// Return a new HttpConfig with the default values filled in
func (c httpConfig) Default() *httpConfig {
	return &httpConfig{
		Port:                     7041,
		ReadTimeout:              15 * time.Second,
		ReadHeaderTimeout:        15 * time.Second,
		WriteTimeout:             15 * time.Second,
		IdleTimeout:              60 * time.Second,
		MaxHeaderBytes:           0,
		WebUiUrl:                 "/ui",
		HttpApiUrl:               "/api",
		WsNegotiationEndpointUrl: "/ws",
	}
}

type webSocketConfig struct {
	AcceptedSubProtocols []string `yaml:"acceptedSubProtocols"`
	InsecureSkipVerify   bool     `yaml:"insecureSkipVerify"`
	OriginPatterns       []string `yaml:"originPatterns"`
	MaxReadLimit         int32    `yaml:"maxReadLimit"`
}

// Return a new webSocketConfig with the default values filled in
func (c webSocketConfig) Default() *webSocketConfig {
	return &webSocketConfig{
		AcceptedSubProtocols: []string{"v12.stomp", "v11.stomp"},
		InsecureSkipVerify:   true,
		OriginPatterns:       nil,
		MaxReadLimit:         math.MaxInt32,
	}
}

type stompConfig struct {
	Login     string        `yaml:"login"`
	Password  string        `yaml:"password"`
	HeartBeat time.Duration `yaml:"heartBeat"`
}

func (c *stompConfig) Authenticate(login, passcode string) bool {
	if c.Login == "" || c.Password == "" {
		return true
	}
	return c.Login == login && c.Password == passcode
}

// Return a new stompConfig with the default values filled in
func (c stompConfig) Default() *stompConfig {
	return &stompConfig{
		Login:     "",
		Password:  "",
		HeartBeat: 15 * time.Second,
	}
}
