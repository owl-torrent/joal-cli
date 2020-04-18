package main

import (
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclient"
	seed2 "github.com/anthonyraymond/joal-cli/pkg/seed"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

// torrent base library : https://github.com/anacrolix/torrent
// especially for bencode and tracker subpackages

func main() {
	var client emulatedclient.EmulatedClient
	clientFile, err := os.Open("C:/Users/raymo/Desktop/joal3/clients/qbittorrent.yml")
	if err != nil {
		panic(err)
	}

	decoder := yaml.NewDecoder(clientFile)
	decoder.SetStrict(true)
	err = decoder.Decode(&client)
	if err != nil {
		panic(err)
	}
	err = client.AfterPropertiesSet()
	if err != nil {
		panic(err)
	}
	err = client.StartListener()
	if err != nil {
		panic(err)
	}

	dispatcher := bandwidth.DispatcherNew(&bandwidth.RandomSpeedProvider{
		MinimumBytesPerSeconds: 0,
		MaximumBytesPerSeconds: 10,
	})

	seed, err := seed2.LoadFromFile(`C:/Users/raymo/Desktop/joal3/torrents/a.torrent`)
	if err != nil {
		panic(err)
	}

	seed.Seed(&client, dispatcher)

	timer := time.NewTimer(10 * time.Second)
	<-timer.C
}
