package announce

type Announcer struct {
	Numwant       int `yaml:"numwant"`
	NumwantOnStop int `yaml:"numwantOnStop"`
	http          IHttpAnnouncer
	udp           IUdpAnnouncer
}
