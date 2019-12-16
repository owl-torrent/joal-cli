package seedmanager

import (
	"bytes"
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/anthonyraymond/joal-cli/pkg/seed"
	assert "github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
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
		if err = os.RemoveAll(testDir); err != nil {
			t.Fatal(err)
		}
	}
}

func createTorrentFile(t *testing.T, directory string) string {
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

	return file.Name()
}

type WaitAbleClient struct {
	wg *sync.WaitGroup
}

func (w *WaitAbleClient) Announce(announceList *metainfo.AnnounceList, infoHash torrent.InfoHash, uploaded int64, downloaded int64, left int64, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
	w.wg.Done()
	return tracker.AnnounceResponse{Interval: 1800}, nil
}
func (w *WaitAbleClient) StartListener() error             { return nil }
func (w *WaitAbleClient) StopListener(ctx context.Context) {}

type DumbDispatcher struct {
	wgStart *sync.WaitGroup
	wgStop  *sync.WaitGroup
}

func (d *DumbDispatcher) Start() {
	if d.wgStart != nil {
		d.wgStart.Done()
	}
}
func (d *DumbDispatcher) Stop() {
	if d.wgStop != nil {
		d.wgStop.Done()
	}
}
func (d *DumbDispatcher) ClaimOrUpdate(claimer bandwidth.IBandwidthClaimable) {}
func (d *DumbDispatcher) Release(claimer bandwidth.IBandwidthClaimable)       {}

func TestSeedManager_Start_ShouldDetectAlreadyPresentFiles(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	announceWg := sync.WaitGroup{}
	manager := &SeedManager{
		client:              &WaitAbleClient{&announceWg},
		bandwidthDispatcher: &DumbDispatcher{},
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 1 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	fileCount := 4
	announceWg.Add(fileCount)

	for i := 0; i < fileCount; i++ {
		createTorrentFile(t, folder)
	}

	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	err = testutils.WaitOrFailAfterTimeout(&announceWg, 5*time.Second)
	if err != nil {
		t.Fatal("timeout reached")
	}
	assert.Len(t, manager.seeds, fileCount)
}

func TestSeedManager_Start_ShouldDetectFileAddition(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	announceWg := sync.WaitGroup{}
	manager := &SeedManager{
		client:              &WaitAbleClient{&announceWg},
		bandwidthDispatcher: &DumbDispatcher{},
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 1 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	manager.torrentFileWatcher.Wait()

	fileCount := 4
	announceWg.Add(fileCount)

	for i := 0; i < fileCount; i++ {
		createTorrentFile(t, folder)
	}

	err = testutils.WaitOrFailAfterTimeout(&announceWg, 5*time.Second)
	if err != nil {
		t.Fatal("timeout reached")
	}
	assert.Len(t, manager.seeds, fileCount)
}

func TestSeedManager_Start_ShouldDetectFileDeletion(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	announceWg := sync.WaitGroup{}
	manager := &SeedManager{
		client:              &WaitAbleClient{&announceWg},
		bandwidthDispatcher: &DumbDispatcher{},
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 1 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	fileCount := 4

	files := make([]string, fileCount)
	for i := 0; i < fileCount; i++ {
		files[i] = createTorrentFile(t, folder)
	}

	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	manager.torrentFileWatcher.Wait()

	announceWg.Add(fileCount)
	err = testutils.WaitOrFailAfterTimeout(&announceWg, 5*time.Second) // wait for creation to be triggered
	if err != nil {
		t.Fatal("timeout reached")
	}

	for _, f := range files {
		announceWg.Add(1)
		err = os.Remove(f)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = testutils.WaitOrFailAfterTimeout(&announceWg, 5*time.Second)
	if err != nil {
		t.Fatal("timeout reached")
	}
	assert.Len(t, manager.seeds, 0)
}

func TestSeedManager_Start_ShouldDetectFileRename(t *testing.T) {
	folder, clean := setupTestFolder(t)
	defer clean()

	announceWg := sync.WaitGroup{}
	manager := &SeedManager{
		client:              &WaitAbleClient{&announceWg},
		bandwidthDispatcher: &DumbDispatcher{},
		joalPaths: &joalPaths{
			torrentFolder: folder,
		},
		seeds:           make(map[torrent.InfoHash]seed.ISeed),
		fileWatcherPoll: 1 * time.Millisecond,
		lock:            &sync.Mutex{},
	}

	fileCount := 1

	file := createTorrentFile(t, folder)

	err := manager.Start()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer manager.torrentFileWatcher.Close()

	manager.torrentFileWatcher.Wait()

	announceWg.Add(fileCount)
	err = testutils.WaitOrFailAfterTimeout(&announceWg, 5*time.Second) // wait for creation to be triggered
	if err != nil {
		t.Fatal("timeout reached")
	}

	announceWg.Add(2)
	err = os.Rename(file, filepath.Join(filepath.Dir(file), "copy-"+filepath.Base(file)))
	if err != nil {
		t.Fatal(err)
	}

	err = testutils.WaitOrFailAfterTimeout(&announceWg, 5*time.Second)
	if err != nil {
		t.Fatal("timeout reached")
	}
	assert.Len(t, manager.seeds, 1)
}

func TestSeedManager_Stop(t *testing.T) {
	t.Fatal("implement")
}
