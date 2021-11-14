package manager2

import (
	"github.com/anthonyraymond/joal-cli/internal/core/torrent2"
)

func RunQueueConsumer(queue *torrent2.AnnounceQueue, announce func(request *torrent2.AnnounceRequest)) {
	for req := range queue.Request() {
		announce(req)
	}
}
