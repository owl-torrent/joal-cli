package domain

import "time"

type AnnounceHistory struct {
	at       time.Time
	interval time.Duration
	seeders  int32
	leechers int32
	error    string
}
