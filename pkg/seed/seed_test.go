package seed

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/mocks"
	"github.com/golang/mock/gomock"
	"github.com/nvn1729/congo"
	"testing"
	"time"
)


func TestSeed_SeedShouldAnnounceInLoopAndUpdateDispatcher(t *testing.T) {
	seed, err := LoadFromFile("./testdata/ubuntu.torrent")
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockIEmulatedClient(ctrl)
	dispatcher := mocks.NewMockIDispatcher(ctrl)
	announceLatch := congo.NewCountDownLatch(3)

	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Started)).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
			_ = announceLatch.CountDown()
			return tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders:  0, Peers: []tracker.Peer{}}, nil
		}).Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.None)).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
			_ = announceLatch.CountDown()
			return tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders:  0, Peers: []tracker.Peer{}}, nil
		}).MinTimes(2)

	dispatcher.
		EXPECT().
		ClaimOrUpdate(gomock.Eq(seed)).
		MinTimes(3)

	go seed.Seed(client, dispatcher)
	announceLatch.WaitTimeout(10 * time.Second)
}


func TestSeed_SeedShouldAnnounceStopOnStopAndReleaseDispatcher(t *testing.T) {
	seed, err := LoadFromFile("./testdata/ubuntu.torrent")
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockIEmulatedClient(ctrl)
	dispatcher := mocks.NewMockIDispatcher(ctrl)
	announceLatch := congo.NewCountDownLatch(3)
	stopLatch := congo.NewCountDownLatch(1)

	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Started)).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
			_ = announceLatch.CountDown()
			return tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders:  0, Peers: []tracker.Peer{}}, nil
		}).Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.None)).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
			_ = announceLatch.CountDown()
			fmt.Println("dd")
			return tracker.AnnounceResponse{Interval: 1, Leechers: 0, Seeders:  0, Peers: []tracker.Peer{}}, nil
		}).AnyTimes()
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Stopped)).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
			_ = stopLatch.CountDown()
			return tracker.AnnounceResponse{Interval: 10, Leechers: 0, Seeders:  0, Peers: []tracker.Peer{}}, nil
		}).Times(1)

	dispatcher.
		EXPECT().
		ClaimOrUpdate(gomock.Eq(seed)).
		MinTimes(3)
	dispatcher.
		EXPECT().
		Release(gomock.Eq(seed)).
		Times(1)

	go seed.Seed(client, dispatcher)
	if !announceLatch.WaitTimeout(10 * time.Second) {
		t.Fatal("announceLatch timed out")
	}
	seed.StopSeeding(context.Background())
	if !stopLatch.WaitTimeout(10 * time.Second) {
		t.Fatal("stopLatch timed out")
	}
}