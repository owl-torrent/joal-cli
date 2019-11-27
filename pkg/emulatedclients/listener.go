package emulatedclients

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type Listener struct {
	Port          Port    `yaml:"port" validate:"required"`
	listeningPort *uint16 `yaml:"-"`
	ip            *net.IP `yaml:"-"`
}
type Port struct {
	Min uint16 `yaml:"min" validate:"min=1"`
	Max uint16 `yaml:"max" validate:"min=1,gtefield=Min"`
}

func (l *Listener) AfterPropertiesSet() error {
	return nil
}

// Blocking call until the listener is ready and public ip is retrieved.
func (l *Listener) Start() error {
	ip, err := getPublicIp()
	if err != nil {
		return err
	}
	l.ip = &ip
	// TODO: Start listening on port for peers requests and answer
	mockedPort := uint16(9000)
	l.listeningPort = &mockedPort
	return nil
}

var publicIpProviders = []string{
	"https://api.ipify.org",
	"http://myexternalip.com/raw",
	"http://ipinfo.io/ip",
	"http://ipecho.net/plain",
	"http://icanhazip.com",
	"http://ifconfig.me/ip",
	"http://ident.me",
	"http://checkip.amazonaws.com",
	"http://bot.whatismyipaddress.com",
	"http://whatismyip.akamai.com",
	"http://wgetip.com",
	"http://ip.appspot.com",
	"http://ip.tyk.nu",
	"https://shtuff.it/myip/short",
}

func getPublicIp() (net.IP, error) {
	for _, providerUri := range publicIpProviders {
		client := &http.Client{Timeout: 8 * time.Second}
		req, err := http.NewRequest("GET", providerUri, nil)
		resp, err := client.Do(req)
		if err != nil {
			// TODO: log error
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			// TODO: log error
			continue
		}

		if resp.StatusCode != 200 {
			//TODO: log error
			continue
		}

		tb := strings.TrimSpace(string(body))
		ip := net.ParseIP(tb)
		if ip == nil {
			// TODO: log error
			continue
		}
		return ip, nil
	}

	return nil, errors.New("failed to get public IP address.")
}