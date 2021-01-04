package web

type WebConfig struct {
	Http  *HttpConfig  `yaml:"http"`
	Stomp *StompConfig `yaml:"stomp"`
}

// Return a new WebConfig with the default values filled in
func (c WebConfig) Default() *WebConfig {
	return &WebConfig{
		Http:  HttpConfig{}.Default(),
		Stomp: StompConfig{}.Default(),
	}
}

type HttpConfig struct {
	Port int `yaml:"port"`
}

// Return a new HttpConfig with the default values filled in
func (c HttpConfig) Default() *HttpConfig {
	return &HttpConfig{
		Port: 5703,
	}
}

type StompConfig struct {
	Port     int    `yaml:"port"`
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
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
		Port:     5704,
		Login:    "",
		Password: "",
	}
}
