package seedmanager

import (
	"bytes"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestFolder(t *testing.T) (absPath string, cleanFunction func()) {
	testDir, err := ioutil.TempDir("./testdata/", "")
	if err != nil {
		t.Fatal(err)
	}

	absPath, err = filepath.Abs(testDir)
	if err != nil {
		_ = os.RemoveAll(testDir)
		t.Fatal(err)
	}
	return absPath, func() {
		for attempts := 0; attempts < 8; attempts++ {
			if err = os.RemoveAll(testDir); err == nil {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
		t.Fatal(err)
	}
}

func createTorrentFile(t *testing.T, directory string) (string, torrent.InfoHash) {
	name := make([]byte, 180)
	rand.Read(name)
	info := metainfo.Info{
		PieceLength: 0,
		Pieces:      []byte{},
		Name:        string(name),
		Length:      0,
	}

	buf := bytes.Buffer{}
	err := bencode.NewEncoder(&buf).Encode(info)
	if err != nil {
		t.Fatal(err)
	}

	meta := metainfo.MetaInfo{
		InfoBytes:    buf.Bytes(),
		Announce:     "http://announce.fr/announce",
		AnnounceList: metainfo.AnnounceList{},
		Nodes:        []metainfo.Node{metainfo.Node("127.0.0.1:1001")},
		CreationDate: 150000,
		Comment:      "forged for test",
		CreatedBy:    "joal test",
	}

	file, err := ioutil.TempFile(directory, "*.torrent")
	if err != nil {
		t.Fatal(err)
	}

	err = meta.Write(file)
	if err != nil {
		_ = file.Close()
		t.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	return file.Name(), meta.HashInfoBytes()
}

/*
func TestSeedManager_Start_ShouldDetectAlreadyPresentFiles(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	ctrl := gomock.NewController(t)
	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)

	manager := &SeedManager{
		client:              client,
		bandwidthDispatcher: dispatcher,
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 5 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	fileCount := uint(4)
	latch := congo.NewCountDownLatch(fileCount)

	for i := uint(0); i < fileCount; i++ {
		_, infoHash := createTorrentFile(t, folder)
		client.
			EXPECT().
			Announce(gomock.Any(), gomock.Eq(infoHash), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(tracker.Started), gomock.Any()).
			DoAndReturn(func(e ...interface{}) (announcer.AnnounceResponse, error) {
				_ = latch.CountDown()
				return announcer.AnnounceResponse{Interval: 1000 * time.Second}, nil
			}).
			Times(1)
	}
	client.EXPECT().StartListener().Times(1)
	dispatcher.EXPECT().Start().AnyTimes()
	dispatcher.EXPECT().ClaimOrUpdate(gomock.Any()).AnyTimes()
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
}*/
/*
func TestSeedManager_Start_ShouldDetectFileAddition(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	ctrl := gomock.NewController(t)
	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)

	manager := &SeedManager{
		client:              client,
		bandwidthDispatcher: dispatcher,
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 5 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	client.EXPECT().StartListener().Times(1)
	dispatcher.EXPECT().Start().AnyTimes()
	dispatcher.EXPECT().ClaimOrUpdate(gomock.Any()).AnyTimes()
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()
	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	manager.torrentFileWatcher.Wait()

	fileCount := uint(4)
	latch := congo.NewCountDownLatch(fileCount)

	for i := uint(0); i < fileCount; i++ {
		_, infoHash := createTorrentFile(t, folder)
		client.
			EXPECT().
			Announce(gomock.Any(), gomock.Eq(infoHash), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(tracker.Started), gomock.Any()).
			DoAndReturn(func(e ...interface{}) (announcer.AnnounceResponse, error) {
				_ = latch.CountDown()
				return announcer.AnnounceResponse{Interval: 1000 * time.Second}, nil
			}).
			Times(1)
	}
	dispatcher.EXPECT().ClaimOrUpdate(gomock.Any()).AnyTimes()
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
}*/
/*
func TestSeedManager_Start_ShouldDetectFileDeletion(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	ctrl := gomock.NewController(t)
	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)

	manager := &SeedManager{
		client:              client,
		bandwidthDispatcher: dispatcher,
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 5 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	fileCount := uint(4)
	latch := congo.NewCountDownLatch(fileCount)

	files := make(map[string]torrent.InfoHash, fileCount)
	for i := uint(0); i < fileCount; i++ {
		file, infoHash := createTorrentFile(t, folder)
		files[file] = infoHash
		client.
			EXPECT().
			Announce(gomock.Any(), gomock.Eq(infoHash), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(tracker.Started), gomock.Any()).
			DoAndReturn(func(e ...interface{}) (announcer.AnnounceResponse, error) {
				_ = latch.CountDown()
				return announcer.AnnounceResponse{Interval: 1000 * time.Second}, nil
			}).
			Times(1)
	}

	client.EXPECT().StartListener().Times(1)
	dispatcher.EXPECT().Start().AnyTimes()
	dispatcher.EXPECT().ClaimOrUpdate(gomock.Any()).AnyTimes()
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	manager.torrentFileWatcher.Wait()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}

	latch = congo.NewCountDownLatch(uint(len(files)))
	for file, infoHash := range files {
		err = os.Remove(file)
		if err != nil {
			t.Fatal(err)
		}
		client.
			EXPECT().
			Announce(gomock.Any(), gomock.Eq(infoHash), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(tracker.Stopped), gomock.Any()).
			DoAndReturn(func(e ...interface{}) (announcer.AnnounceResponse, error) {
				_ = latch.CountDown()
				return announcer.AnnounceResponse{Interval: 1000 * time.Second}, nil
			}).
			Times(1)
	}

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
}*/
/*
func TestSeedManager_Start_ShouldDetectFileRename(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	ctrl := gomock.NewController(t)
	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)

	manager := &SeedManager{
		client:              client,
		bandwidthDispatcher: dispatcher,
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 5 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	latch := congo.NewCountDownLatch(1)

	file, infoHash := createTorrentFile(t, folder)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Eq(infoHash), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(tracker.Started), gomock.Any()).
		DoAndReturn(func(e ...interface{}) (announcer.AnnounceResponse, error) {
			_ = latch.CountDown()
			return announcer.AnnounceResponse{Interval: 1000 * time.Second}, nil
		}).
		Times(1)

	client.EXPECT().StartListener().Times(1)
	dispatcher.EXPECT().Start().AnyTimes()
	dispatcher.EXPECT().ClaimOrUpdate(gomock.Any()).AnyTimes()
	dispatcher.EXPECT().Release(gomock.Any()).AnyTimes()

	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	manager.torrentFileWatcher.Wait()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}

	latch = congo.NewCountDownLatch(2)
	err = os.Rename(file, filepath.Join(filepath.Dir(file), "copy-"+filepath.Base(file)))
	if err != nil {
		t.Fatal(err)
	}
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Eq(infoHash), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(tracker.Stopped), gomock.Any()).
		DoAndReturn(func(e ...interface{}) (announcer.AnnounceResponse, error) {
			_ = latch.CountDown()
			return announcer.AnnounceResponse{Interval: 1000 * time.Second}, nil
		}).
		Times(1)
	client.
		EXPECT().
		Announce(gomock.Any(), gomock.Eq(infoHash), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(int64(0)), gomock.Eq(tracker.Started), gomock.Any()).
		DoAndReturn(func(e ...interface{}) (announcer.AnnounceResponse, error) {
			_ = latch.CountDown()
			return announcer.AnnounceResponse{Interval: 1000 * time.Second}, nil
		}).
		Times(1)

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
}*/
/*
func TestSeedManager_StartAndStop(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	ctrl := gomock.NewController(t)
	client := emulatedclient.NewMockIEmulatedClient(ctrl)
	dispatcher := bandwidth.NewMockIDispatcher(ctrl)

	manager := &SeedManager{
		client:              client,
		bandwidthDispatcher: dispatcher,
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 5 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	listenerLatch := congo.NewCountDownLatch(1)
	dispatcherLatch := congo.NewCountDownLatch(1)
	client.EXPECT().StartListener().Do(func() { _ = listenerLatch.CountDown() }).Times(1)
	dispatcher.EXPECT().Start().Do(func() { _ = dispatcherLatch.CountDown() }).Times(1)

	err := manager.Start()
	if err != nil {
		t.Fatal(err)
	}

	if !listenerLatch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
	if !dispatcherLatch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}

	seeds := make([]seed.ISeed, 255)
	stopSeedLatch := congo.NewCountDownLatch(255)
	for i := 0; i < 255; i++ {
		s := seed.NewMockISeed(ctrl)
		seeds[i] = s
		manager.seeds[[20]byte{byte(i)}] = s

		s.
			EXPECT().
			StopSeeding(gomock.Any()).
			Do(func(e interface{}) { _ = stopSeedLatch.CountDown() }).
			Times(1)
	}
	dispatcher.
		EXPECT().
		Release(gomock.Any()).
		Do(func(e interface{}) { _ = dispatcherLatch.CountDown() }).
		Times(255)
	dispatcher.
		EXPECT().
		Stop().
		Do(func() { _ = dispatcherLatch.CountDown() }).
		Times(1)
	client.
		EXPECT().
		StopListener(gomock.Any()).
		Do(func(e interface{}) { _ = listenerLatch.CountDown() }).
		Times(1)

	dispatcherLatch = congo.NewCountDownLatch(1)
	listenerLatch = congo.NewCountDownLatch(1)

	manager.Stop(context.Background())

	if !stopSeedLatch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
	if !listenerLatch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
	if !dispatcherLatch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch timed out")
	}
}*/
