package torrent

//go:generate pegomock generate --use-experimental-model-gen --package=torrent --self_package=github.com/anthonyraymond/joal-cli/pkg/torrent --output=orchestrator_mock.go github.com/anthonyraymond/joal-cli/pkg/torrent Orchestrator
//go:generate pegomock generate --use-experimental-model-gen --package=torrent --self_package=github.com/anthonyraymond/joal-cli/pkg/torrent --output=tier_mock.go github.com/anthonyraymond/joal-cli/pkg/torrent ITierAnnouncer
//go:generate pegomock generate --use-experimental-model-gen --package=torrent --self_package=github.com/anthonyraymond/joal-cli/pkg/torrent --output=tracker_mock.go github.com/anthonyraymond/joal-cli/pkg/torrent ITrackerAnnouncer

/*
TODO: when https://github.com/petergtz/pegomock/pull/104 will get merged and publish update the global deps to and replace the three generate above with the below one
go:generate pegomock generate --use-experimental-model-gen --package=torrent --self_package=github.com/anthonyraymond/joal-cli/pkg/torrent --output=orchestrator_mock.go orchestrator.go
*/

import (
	"github.com/onsi/ginkgo"
	//"github.com/petergtz/pegomock"
)

var _ = ginkgo.Describe("AllTierAnnouncer", func() {

	/*var (
		tier ITierAnnouncer
		trackers []ITrackerAnnouncer
	)

	ginkgo.BeforeEach(func() {
		trackers = []ITrackerAnnouncer{
			pegomock.Moc
		}
	})*/
})
