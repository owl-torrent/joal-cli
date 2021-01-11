package web

import (
	"math"
	"time"
)

type WebConfig struct {
	Http      *HttpConfig      `yaml:"http"`
	WebSocket *WebSocketConfig `yaml:"webSocket"`
	Stomp     *StompConfig     `yaml:"stomp"`
}

// Return a new WebConfig with the default values filled in
func (c WebConfig) Default() *WebConfig {
	return &WebConfig{
		Http:      HttpConfig{}.Default(),
		WebSocket: WebSocketConfig{}.Default(),
		Stomp:     StompConfig{}.Default(),
	}
}

type HttpConfig struct {
	Port                int           `yaml:"port"`
	ReadTimeout         time.Duration `yaml:"readTimeout"`
	ReadHeaderTimeout   time.Duration `yaml:"readHeaderTimeout"`
	WriteTimeout        time.Duration `yaml:"writeTimeout"`
	IdleTimeout         time.Duration `yaml:"idleTimeout"`
	MaxHeaderBytes      int           `yaml:"maxHeaderBytes"`
	WebUiPath           string        `yaml:"webUiPath"`
}

// Return a new HttpConfig with the default values filled in
func (c HttpConfig) Default() *HttpConfig {
	return &HttpConfig{
		Port:              7041,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		WebUiPath:         "/ui",
	}
}

type WebSocketConfig struct {
	AcceptedSubProtocols []string `yaml:"acceptedSubProtocols"`
	InsecureSkipVerify   bool     `yaml:"insecureSkipVerify"`
	OriginPatterns       []string `yaml:"originPatterns"`
	MaxReadLimit         int32    `yaml:"maxReadLimit"`
}

// Return a new HttpConfig with the default values filled in
func (c WebSocketConfig) Default() *WebSocketConfig {
	return &WebSocketConfig{
		AcceptedSubProtocols: []string{"v12.stomp", "v11.stomp"},
		InsecureSkipVerify:   true,
		OriginPatterns:       nil,
		MaxReadLimit:         math.MaxInt32,
	}
}

type StompConfig struct {
	Login     string `yaml:"login"`
	Password  string `yaml:"password"`
	UrlPath   string `yaml:"urlPath"`
	HeartBeat time.Duration
}

func (c *StompConfig) Authenticate(login, passcode string) bool {
	if c.Login == "" || c.Password == "" {
		return true
	}
	return c.Login == login && c.Password == passcode
}

// Return a new StompConfig with the default values filled in
func (c StompConfig) Default() *StompConfig {
	return &StompConfig{
		Login:     "",
		Password:  "",
		UrlPath:   "/ws",
		HeartBeat: 15 * time.Second,
	}
}
