package seed

import (
	"fmt"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func createTorrentFile(t *testing.T, directory string, metaAdapters ...func(info *metainfo.Info, meta *metainfo.MetaInfo)) (string, metainfo.MetaInfo) {
	meta := &metainfo.MetaInfo{
		Announce: "http://localhost:8080/announce",
		AnnounceList: metainfo.AnnounceList{
			[]string{"http://localhost:8080/announce", "http://localhost:9090/announce", "http://localhost:6060/announce"},
			[]string{"http://localhost:3030/announce", "http://localhost:2020/announce", "http://localhost:1010/announce"},
		},
		Nodes:        []metainfo.Node{metainfo.Node("127.0.0.1:1001")},
		CreationDate: 150000,
		Comment:      "forged for test",
		CreatedBy:    "me",
	}

	name := make([]byte, 180)
	rand.Read(name)
	info := &metainfo.Info{
		PieceLength: 2,
		Pieces:      []byte{0x01, 0x02},
		Name:        string(name),
		Length:      0,
	}
	infoBytes, err := bencode.Marshal(*info)
	if err != nil {
		t.Fatal(err)
	}

	for _, adapt := range metaAdapters {
		adapt(info, meta)
	}
	info.PieceLength = int64(len(info.Pieces))
	meta.InfoBytes = infoBytes

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

	return file.Name(), *meta
}

func Test_Torrent_ShouldReadFromFile(t *testing.T) {
	tempDir := t.TempDir()

	torrentFile, expectedMeta := createTorrentFile(t, tempDir)
	actualTorrent, err := FromFile(torrentFile)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expectedMeta.HashInfoBytes().Bytes(), actualTorrent.InfoHash().Bytes())
	assert.Equal(t, expectedMeta.HashInfoBytes().Bytes(), actualTorrent.InfoHash().Bytes())
	actualTorrentAsJoalTorrent := actualTorrent.(*joalTorrent)
	assert.Equal(t, torrentFile, actualTorrentAsJoalTorrent.path)
	assert.Equal(t, expectedMeta.Announce, actualTorrentAsJoalTorrent.metaInfo.Announce)
	expectedAnnounceList := expectedMeta.AnnounceList.Clone()
	for _, tier := range expectedAnnounceList {
		sort.Slice(tier, func(i, j int) bool {
			return tier[i] < tier[j]
		})
	}
	actualAnnounceList := actualTorrentAsJoalTorrent.metaInfo.AnnounceList.Clone()
	for _, tier := range actualAnnounceList {
		sort.Slice(tier, func(i, j int) bool {
			return tier[i] < tier[j]
		})
	}
	assert.Equal(t, expectedAnnounceList, actualAnnounceList)
	assert.Equal(t, expectedMeta.Comment, actualTorrentAsJoalTorrent.metaInfo.Comment)
	assert.Equal(t, expectedMeta.CreatedBy, actualTorrentAsJoalTorrent.metaInfo.CreatedBy)
	assert.Equal(t, expectedMeta.CreationDate, actualTorrentAsJoalTorrent.metaInfo.CreationDate)
	assert.Equal(t, expectedMeta.Encoding, actualTorrentAsJoalTorrent.metaInfo.Encoding)
	assert.Equal(t, expectedMeta.Nodes, actualTorrentAsJoalTorrent.metaInfo.Nodes)
	assert.Equal(t, expectedMeta.UrlList, actualTorrentAsJoalTorrent.metaInfo.UrlList)
	expectedInfo, err := expectedMeta.UnmarshalInfo()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedInfo.Files, actualTorrentAsJoalTorrent.info.Files)
	assert.Equal(t, expectedInfo.Length, actualTorrentAsJoalTorrent.info.Length)
	assert.Equal(t, expectedInfo.Name, actualTorrentAsJoalTorrent.info.Name)
	assert.Equal(t, expectedInfo.PieceLength, actualTorrentAsJoalTorrent.info.PieceLength)
	assert.Equal(t, expectedInfo.Private, actualTorrentAsJoalTorrent.info.Private)
	assert.Equal(t, expectedInfo.Source, actualTorrentAsJoalTorrent.info.Source)
}

func Test_Torrent_ShouldShuffleTrackerInTiers(t *testing.T) {
	tempDir := t.TempDir()

	torrentFile, expectedMeta := createTorrentFile(t, tempDir, func(info *metainfo.Info, meta *metainfo.MetaInfo) {
		meta.AnnounceList = metainfo.AnnounceList{
			[]string{"http://localhost:1000/announce", "http://localhost:1001/announce", "http://localhost:1002/announce", "http://localhost:1003/announce", "http://localhost:1004/announce", "http://localhost:1005/announce", "http://localhost:1006/announce", "http://localhost:1007/announce", "http://localhost:1008/announce", "http://localhost:1009/announce", "http://localhost:1010/announce", "http://localhost:1011/announce", "http://localhost:1012/announce", "http://localhost:1013/announce", "http://localhost:1014/announce", "http://localhost:1015/announce", "http://localhost:1016/announce", "http://localhost:1017/announce", "http://localhost:1018/announce", "http://localhost:1019/announce"},
			[]string{"http://localhost:2000/announce", "http://localhost:2001/announce", "http://localhost:2002/announce", "http://localhost:2003/announce", "http://localhost:2004/announce", "http://localhost:2005/announce", "http://localhost:2006/announce", "http://localhost:2007/announce", "http://localhost:2008/announce", "http://localhost:2009/announce", "http://localhost:2010/announce", "http://localhost:2011/announce", "http://localhost:2012/announce", "http://localhost:2013/announce", "http://localhost:2014/announce", "http://localhost:2015/announce", "http://localhost:2016/announce", "http://localhost:2017/announce", "http://localhost:2018/announce", "http://localhost:2019/announce"},
			[]string{"http://localhost:3000/announce", "http://localhost:3001/announce", "http://localhost:3002/announce", "http://localhost:3003/announce", "http://localhost:3004/announce", "http://localhost:3005/announce", "http://localhost:3006/announce", "http://localhost:3007/announce", "http://localhost:3008/announce", "http://localhost:3009/announce", "http://localhost:3010/announce", "http://localhost:3011/announce", "http://localhost:3012/announce", "http://localhost:3013/announce", "http://localhost:3014/announce", "http://localhost:3015/announce", "http://localhost:3016/announce", "http://localhost:3017/announce", "http://localhost:3018/announce", "http://localhost:3019/announce"},
			[]string{"http://localhost:4000/announce", "http://localhost:4001/announce", "http://localhost:4002/announce", "http://localhost:4003/announce", "http://localhost:4004/announce", "http://localhost:4005/announce", "http://localhost:4006/announce", "http://localhost:4007/announce", "http://localhost:4008/announce", "http://localhost:4009/announce", "http://localhost:4010/announce", "http://localhost:4011/announce", "http://localhost:4012/announce", "http://localhost:4013/announce", "http://localhost:4014/announce", "http://localhost:4015/announce", "http://localhost:4016/announce", "http://localhost:4017/announce", "http://localhost:4018/announce", "http://localhost:4019/announce"},
			[]string{"http://localhost:5000/announce", "http://localhost:5001/announce", "http://localhost:5002/announce", "http://localhost:5003/announce", "http://localhost:5004/announce", "http://localhost:5005/announce", "http://localhost:5006/announce", "http://localhost:5007/announce", "http://localhost:5008/announce", "http://localhost:5009/announce", "http://localhost:5010/announce", "http://localhost:5011/announce", "http://localhost:5012/announce", "http://localhost:5013/announce", "http://localhost:5014/announce", "http://localhost:5015/announce", "http://localhost:5016/announce", "http://localhost:5017/announce", "http://localhost:5018/announce", "http://localhost:5019/announce"},
			[]string{"http://localhost:6000/announce", "http://localhost:6001/announce", "http://localhost:6002/announce", "http://localhost:6003/announce", "http://localhost:6004/announce", "http://localhost:6005/announce", "http://localhost:6006/announce", "http://localhost:6007/announce", "http://localhost:6008/announce", "http://localhost:6009/announce", "http://localhost:6010/announce", "http://localhost:6011/announce", "http://localhost:6012/announce", "http://localhost:6013/announce", "http://localhost:6014/announce", "http://localhost:6015/announce", "http://localhost:6016/announce", "http://localhost:6017/announce", "http://localhost:6018/announce", "http://localhost:6019/announce"},
			[]string{"http://localhost:7000/announce", "http://localhost:7001/announce", "http://localhost:7002/announce", "http://localhost:7003/announce", "http://localhost:7004/announce", "http://localhost:7005/announce", "http://localhost:7006/announce", "http://localhost:7007/announce", "http://localhost:7008/announce", "http://localhost:7009/announce", "http://localhost:7010/announce", "http://localhost:7011/announce", "http://localhost:7012/announce", "http://localhost:7013/announce", "http://localhost:7014/announce", "http://localhost:7015/announce", "http://localhost:7016/announce", "http://localhost:7017/announce", "http://localhost:7018/announce", "http://localhost:7019/announce"},
			[]string{"http://localhost:8000/announce", "http://localhost:8001/announce", "http://localhost:8002/announce", "http://localhost:8003/announce", "http://localhost:8004/announce", "http://localhost:8005/announce", "http://localhost:8006/announce", "http://localhost:8007/announce", "http://localhost:8008/announce", "http://localhost:8009/announce", "http://localhost:8010/announce", "http://localhost:8011/announce", "http://localhost:8012/announce", "http://localhost:8013/announce", "http://localhost:8014/announce", "http://localhost:8015/announce", "http://localhost:8016/announce", "http://localhost:8017/announce", "http://localhost:8018/announce", "http://localhost:8019/announce"},
		}
	})

	reduceTrackerUrls := func(al metainfo.AnnounceList) []string {
		var res []string
		for _, tier := range al {
			res = append(res, tier...)
		}
		return res
	}

	expectedUrls := reduceTrackerUrls(expectedMeta.AnnounceList)
	atLeastOneIsDifferentFromOriginal := false
	for i := 0; i < 10; i++ { // it's all base on rand, out of a bad luck the shuffle may result in exactly the same order as the input. Run multiple time for more robustness
		actualTorrent, err := FromFile(torrentFile)
		if err != nil {
			t.Fatal(err)
		}

		// Should not shuffle tiers
		for i := 0; i < len(expectedMeta.AnnounceList); i++ {
			assert.Contains(t, actualTorrent.(*joalTorrent).metaInfo.AnnounceList[i][0], fmt.Sprintf("localhost:%d0", i+1))
		}

		// Should shuffle trackers in tiers
		if !reflect.DeepEqual(expectedUrls, reduceTrackerUrls(actualTorrent.(*joalTorrent).metaInfo.AnnounceList)) {
			atLeastOneIsDifferentFromOriginal = true
			break
		}
	}
	assert.True(t, atLeastOneIsDifferentFromOriginal)
}

func Test_JoalTorrent_ShouldRegisterTorrentsToBandwidthDispatcherOnAnnounce(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_JoalTorrent_ShouldUnRegisterTorrentsFromBandwidthDispatcherOnStop(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_JoalTorrent_ShouldStartOrchestratorOnStartAndStopOnStop(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_JoalTorrent_ShouldNotifySwarmOnAnnyAnnounceResponseButStopped(t *testing.T) {
	t.Fatal("not implemented")
}
