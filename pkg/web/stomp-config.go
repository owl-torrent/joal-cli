package web

type StompConfig struct {
	Port     int    `json:"port"`
	Login    string `json:"login"`
	Password string `json:"password"`
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
