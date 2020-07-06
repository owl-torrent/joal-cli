package seed

type ISwarm interface {
	GetSeeders() int32
	GetLeechers() int32
}

type swarm struct {
	seeders  int32
	leechers int32
}

func (s *swarm) GetSeeders() int32 {
	return s.seeders
}
func (s *swarm) GetLeechers() int32 {
	return s.leechers
}
