package seed

/*
type BandwidthClaimableSwarmMatcher struct {
	leechers int32
	seeders  int32
}

func (m *BandwidthClaimableSwarmMatcher) Matches(x interface{}) bool {
	swarm, ok := x.(bandwidth.IBandwidthClaimable)
	if !ok {
		return false
	}

	if swarm.GetSwarm().GetSeeders() != m.seeders {
		return false
	}
	if swarm.GetSwarm().GetLeechers() != m.leechers {
		return false
	}
	return true
}
func (m *BandwidthClaimableSwarmMatcher) String() string {
	return fmt.Sprintf("has %d leechers and %d seeders", m.leechers, m.seeders)
}

func TestSeed_SeedShouldAnnounceInLoopAndUpdateDispatcher(t *testing.T) {
	seed, err := LoadFromFile("./testdata/ubuntu.torrent")
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)
	claimLatch := congo.NewCountDownLatch(3)

	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Started), gomock.Any()).
		Return(tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders: 0, Peers: []tracker.Peer{}}, nil).
		Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.None), gomock.Any()).
		Return(tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders: 0, Peers: []tracker.Peer{}}, nil).
		MinTimes(2)

	dispatcher.
		EXPECT().
		ClaimOrUpdate(gomock.Eq(seed)).
		Do(func(e interface{}) { _ = claimLatch.CountDown() }).
		MinTimes(3)
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	go seed.Seed(client, dispatcher)
	if !claimLatch.WaitTimeout(10 * time.Second) {
		t.Fatal("latch has timed out")
	}
}

func TestSeed_SeedShouldAnnounceStopOnStopAndReleaseDispatcher(t *testing.T) {
	seed, err := LoadFromFile("./testdata/ubuntu.torrent")
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)
	announceLatch := congo.NewCountDownLatch(3)
	stopLatch := congo.NewCountDownLatch(1)

	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Started), gomock.Any()).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
			_ = announceLatch.CountDown()
			return tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders: 0, Peers: []tracker.Peer{}}, nil
		}).Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.None), gomock.Any()).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
			_ = announceLatch.CountDown()
			return tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders: 0, Peers: []tracker.Peer{}}, nil
		}).AnyTimes()
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Stopped), gomock.Any()).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
			_ = stopLatch.CountDown()
			return tracker.AnnounceResponse{Interval: 10, Leechers: 0, Seeders: 0, Peers: []tracker.Peer{}}, nil
		}).Times(1)

	dispatcher.
		EXPECT().
		ClaimOrUpdate(gomock.Eq(seed)).
		AnyTimes()
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

func TestSeed_SeedShouldUpdateSeedSwarmWithAnnounceResponse(t *testing.T) {
	seed, err := LoadFromFile("./testdata/ubuntu.torrent")
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)
	claimLatch := congo.NewCountDownLatch(3)

	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Started), gomock.Any()).
		Return(tracker.AnnounceResponse{Interval: 0, Leechers: 10, Seeders: 501, Peers: []tracker.Peer{}}, nil).
		Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.None), gomock.Any()).
		Return(tracker.AnnounceResponse{Interval: 0, Leechers: 0, Seeders: 0, Peers: []tracker.Peer{}}, nil).
		Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.None), gomock.Any()).
		Return(tracker.AnnounceResponse{Interval: 0, Leechers: 20, Seeders: 68, Peers: []tracker.Peer{}}, nil).
		AnyTimes()

	gomock.InOrder(
		dispatcher.EXPECT().ClaimOrUpdate(&BandwidthClaimableSwarmMatcher{leechers: 10, seeders: 500}).Do(func(e interface{}) { _ = claimLatch.CountDown() }).Times(1),
		dispatcher.EXPECT().ClaimOrUpdate(&BandwidthClaimableSwarmMatcher{leechers: 0, seeders: 0}).Do(func(e interface{}) { _ = claimLatch.CountDown() }).Times(1),
		dispatcher.EXPECT().ClaimOrUpdate(&BandwidthClaimableSwarmMatcher{leechers: 20, seeders: 67}).Do(func(e interface{}) { _ = claimLatch.CountDown() }).Times(1),
		dispatcher.EXPECT().ClaimOrUpdate(gomock.Any()).AnyTimes(),
	)
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	go seed.Seed(client, dispatcher)
	if !claimLatch.WaitTimeout(10 * time.Second) {
		t.Fatal("latch has timed out")
	}
}

func TestSeed_SeedResetSwarmWhenAnnounceErrorMoreThanTwice(t *testing.T) {
	seed, err := LoadFromFile("./testdata/ubuntu.torrent")
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)
	announceLatch := congo.NewCountDownLatch(3)
	claimLatch := congo.NewCountDownLatch(2)

	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Started), gomock.Any()).
		Return(tracker.AnnounceResponse{Interval: 0, Leechers: 10, Seeders: 501, Peers: []tracker.Peer{}}, nil).
		Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.None), gomock.Any()).
		Return(tracker.AnnounceResponse{}, errors.New("emulate an error")).
		MinTimes(2)

	gomock.InOrder(
		dispatcher.EXPECT().ClaimOrUpdate(&BandwidthClaimableSwarmMatcher{leechers: 10, seeders: 500}).Do(func(e interface{}) { _ = claimLatch.CountDown() }).Times(1),
		dispatcher.EXPECT().ClaimOrUpdate(&BandwidthClaimableSwarmMatcher{leechers: 0, seeders: 0}).Do(func(e interface{}) { _ = claimLatch.CountDown() }).MinTimes(1),
	)
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	go seed.Seed(client, dispatcher)
	if !claimLatch.WaitTimeout(10 * time.Second) {
		fmt.Println(announceLatch.Count())
		t.Fatal("latch has timed out")
	}
}

func TestSeed_SeedShouldNotFailIfAnnounceStartedIsAnError(t *testing.T) {
	seed, err := LoadFromFile("./testdata/ubuntu.torrent")
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)
	announceLatch := congo.NewCountDownLatch(1)

	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(tracker.Started), gomock.Any()).
		DoAndReturn(func(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
			_ = announceLatch.CountDown()
			return tracker.AnnounceResponse{}, errors.New("emulate an error")
		}).Times(1)

	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	go seed.Seed(client, dispatcher)
	if !announceLatch.WaitTimeout(10 * time.Second) {
		fmt.Println(announceLatch.Count())
		t.Fatal("latch has timed out")
	}
}
*/
