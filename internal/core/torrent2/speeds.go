package torrent2

import "sync"

type Speeds interface {
	UploadSpeed() (bps int64)
	SetUploadSpeed(bps int64)
	Reset()
}

type speedsImpl struct {
	upload int64
	lock   *sync.Mutex
}

func newSpeed() Speeds {
	return &speedsImpl{
		upload: 0,
		lock:   &sync.Mutex{},
	}
}

func (s *speedsImpl) UploadSpeed() int64 {
	s.lock.Lock()
	v := s.upload
	s.lock.Unlock()
	return v
}

func (s *speedsImpl) SetUploadSpeed(bps int64) {
	s.lock.Lock()
	s.upload = bps
	s.lock.Unlock()
}

func (s *speedsImpl) Reset() {
	s.upload = 0
}
