package seedmanager

// prend un objet JoalConfig (une struct qui contient la liste des fichiers clients (et leurs path), la liste des torrents, la liste des torrents archivés, la runtime config etc...

// gere les mouvement dans les dossiers

type ITorrentManager interface {
	Seed() error
	StopSeeding()
}

type torrentManager struct {
}

func (t torrentManager) Seed() error {
	panic("implement me")
}

func (t torrentManager) StopSeeding() {
	panic("implement me")
}

func NewTorrentManager() ITorrentManager {
	return &torrentManager{}
}
