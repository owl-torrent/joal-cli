package torrent2

import (
	"sync"
)

// Stats stores the torrent current uploaded, downloaded, left and corrupted stats. Implementation MUST be thread-safe
type Stats interface {
	Uploaded() int64
	Downloaded() int64
	Left() int64
	Corrupted() int64
	AddUploaded(int64)
	AddDownloaded(int64)
	Reset()
}

type statsImpl struct {
	uploaded   int64
	downloaded int64
	left       int64
	corrupted  int64
	lock       *sync.Mutex
}

func newStats() Stats {
	return &statsImpl{
		uploaded:   0,
		downloaded: 0,
		left:       0,
		corrupted:  0,
		lock:       &sync.Mutex{},
	}
}

func (s *statsImpl) Uploaded() int64 {
	s.lock.Lock()
	v := s.uploaded
	s.lock.Unlock()
	return v
}

func (s *statsImpl) Downloaded() int64 {
	s.lock.Lock()
	v := s.downloaded
	s.lock.Unlock()
	return v
}

func (s *statsImpl) Left() int64 {
	s.lock.Lock()
	v := s.left
	s.lock.Unlock()
	return v
}

func (s *statsImpl) Corrupted() int64 {
	s.lock.Lock()
	v := s.corrupted
	s.lock.Unlock()
	return v
}

func (s *statsImpl) AddUploaded(uploaded int64) {
	s.lock.Lock()
	s.uploaded += uploaded
	s.lock.Unlock()
}

func (s *statsImpl) AddDownloaded(downloaded int64) {
	s.lock.Lock()
	s.downloaded += downloaded
	s.lock.Unlock()
	// TODO: decrement LEFT
	// TODO: update corrupted once in a while?
}

func (s *statsImpl) Reset() {
	s.uploaded = 0
	s.downloaded = 0
	s.left = 0
	s.corrupted = 0
}
