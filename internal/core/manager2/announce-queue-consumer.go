package manager2

import (
	"github.com/anthonyraymond/joal-cli/internal/core/announces"
	"github.com/anthonyraymond/joal-cli/internal/core/torrent2"
)

func RunQueueConsumer(queue *torrent2.AnnounceQueue, announce func(request *announces.AnnounceRequest)) {
	for req := range queue.Request() {
		announce(req)
	}
}
