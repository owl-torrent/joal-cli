package seedmanager

import "context"

// prend un objet JoalConfig (une struct qui contient la liste des fichiers clients (et leurs path), la liste des torrents, la liste des torrents archivés, la runtime config etc...

// gere les mouvement dans les dossiers

type ITorrentManager interface {
	Seed() error
	StopSeeding(ctx context.Context)
}

type torrentManager struct {
	//JoalConfigLoader => pas encore prêt, pour le moment on ne l'utilise pas
}

func (t torrentManager) Seed() error {
	// pour le moment on instancie une JoalConfig, (après on utilisera le NewJoalConfigLoader mais il n'est pas prêt), donc créer en une à la main

	// Si le JoalConfig.Client est vide on va prendre le premier client dispo ( dans le dossier config.ClientsDir) puis on créer un EmulatedClient avec, sinon on prend le JoalConfig.Client pour créer le client
	// Un exemple de client est dispo dans pkg/emulatedclient/testdata/client.yml
	// créer un dispatcher avec JoalConfig.RuntimeConfig
	// start du dispatcher

	//torrents := make(map[string]seed.ITorrent)
	// Ensuite on va créer un watcher "github.com/anthonyraymond/watcher" (voir example dans pkg/seedmanager/seed-manager.go), c'est lui qui va déclencher des evenement quand il y a des création/suppression de torrent dans les dossiers
	//   en cas de création de torrent => on instiancie le torrent et on le StartSeeding()
	//   en cas de suppression => on stop le torrent et on supprime de la map des torrents
	//   en cas de rename, ben on rename...

	// il faut aussi que les fichiers déja présent se démarrent en seed. (il y a un exemple dans pkg/seedmanager/seed-manager.go aussi)

	// démarage du emulatedCLient.StartListener(), celui qui va choke tout le monde

	return nil
}

func (t torrentManager) StopSeeding(ctx context.Context) {
	// on envoi un signal d'arrêt a la goroutine qui seed voir exemple de Start & Stop dans pkg/seed/torent.go
}

func NewTorrentManager() ITorrentManager {
	return &torrentManager{
		//joalConfigLoader: ...
	}
}
