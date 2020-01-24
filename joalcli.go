package main

// torrent base library : https://github.com/anacrolix/torrent
// especially for bencode and tracker subpackages

func main() {
	/*var client emulatedclient.EmulatedClient
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

	seed, err := torrent.LoadFromFile(`C:/Users/raymo/Desktop/joal3/torrents/a.torrent`, &client)
	if err != nil {
		panic(err)
	}

	seed.Seed()

	timer := time.NewTimer(10 * time.Second)
	<-timer.C*/
}
