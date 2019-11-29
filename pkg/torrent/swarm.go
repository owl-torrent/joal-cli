package torrent

type swarm struct {
	seeders  uint16
	leechers uint16
}

func (s *swarm) GetSeeders() uint16 {
	return s.seeders
}
func (s *swarm) GetLeechers() uint16 {
	return s.leechers
}
