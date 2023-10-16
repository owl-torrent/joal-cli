package domain

import (
	commonDomain "github.com/anthonyraymond/joal-cli/internal/old/v3/commons/domain"
	"net/url"
)

type Tracker struct {
	Url     url.URL
	Tier    int
	State   TrackerState
	History []AnnounceHistory
	Peers   commonDomain.Peers
}
