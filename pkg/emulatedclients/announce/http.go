package announce

import (
	"bytes"
	"fmt"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/missinggo/httptoo"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/tracker"
	"io"
	"net"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

type httpResponse struct {
	FailureReason string `bencode:"failure reason"`
	Interval      int32  `bencode:"interval"`
	TrackerId     string `bencode:"tracker id"`
	Complete      int32  `bencode:"complete"`
	Incomplete    int32  `bencode:"incomplete"`
	Peers         Peers  `bencode:"peers"`
	// BEP 7
	Peers6 krpc.CompactIPv6NodeAddrs `bencode:"peers6"`
}

type IHttpAnnouncer interface {
	Announce(url url.URL, announceRequest tracker.AnnounceRequest) (tracker.AnnounceResponse, error)
	AfterPropertiesSet() error
}

type HttpAnnouncer struct {
	ShouldUrlEncode bool                `yaml:"shouldUrlEncode"`
	Query           string              `yaml:"string"`
	RequestHeaders  []HttpRequestHeader `yaml:"requestHeaders"`
	queryTemplate   *template.Template  `yaml:"-"`
}

func (a *HttpAnnouncer) AfterPropertiesSet() error {
	var err error
	a.queryTemplate, err = template.New("httpQueryTemplate").Funcs(templateFunctions).Parse(a.Query)
	if err != nil {
		return err
	}
	return nil
}

func (a *HttpAnnouncer) Announce(url url.URL, announceRequest tracker.AnnounceRequest) (ret tracker.AnnounceResponse, err error) {
	_url := httptoo.CopyURL(&url)
	setupQuery(_url, announceRequest)
	req, err := http.NewRequest("GET", _url.String(), nil)
	/*if opt.Context != nil {
		req = req.WithContext(opt.Context)
	}*/
	resp, err := (&http.Client{
		Timeout: time.Second * 15,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 15 * time.Second,
			}).DialContext,
			//Proxy:               opt.HTTPProxy,
			TLSHandshakeTimeout: 15 * time.Second,
			/*TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         opt.ServerName,
			},*/
		},
	}).Do(req)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("response from tracker: %s: %s", resp.Status, buf.String())
		return
	}
	var trackerResponse httpResponse
	err = bencode.Unmarshal(buf.Bytes(), &trackerResponse)
	if _, ok := err.(bencode.ErrUnusedTrailingBytes); ok {
		err = nil
	} else if err != nil {
		err = fmt.Errorf("error decoding %q: %s", buf.Bytes(), err)
		return
	}
	if trackerResponse.FailureReason != "" {
		err = fmt.Errorf("tracker gave failure reason: %q", trackerResponse.FailureReason)
		return
	}
	ret.Interval = trackerResponse.Interval
	ret.Leechers = trackerResponse.Incomplete
	ret.Seeders = trackerResponse.Complete
	ret.Peers = trackerResponse.Peers
	for _, na := range trackerResponse.Peers6 {
		ret.Peers = append(ret.Peers, tracker.Peer{
			IP:   na.IP,
			Port: na.Port,
		})
	}
	return
}

type HttpRequestHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Peers []tracker.Peer

func (me *Peers) UnmarshalBencode(b []byte) (err error) {
	var _v interface{}
	err = bencode.Unmarshal(b, &_v)
	if err != nil {
		return
	}
	switch v := _v.(type) {
	case string:
		//FIXME: vars.Add("http responses with string peers", 1)
		var cnas krpc.CompactIPv4NodeAddrs
		err = cnas.UnmarshalBinary([]byte(v))
		if err != nil {
			return
		}
		for _, cp := range cnas {
			localP := Peer{
				IP:   cp.IP[:],
				Port: cp.Port,
			}
			*me = append(*me, tracker.Peer{
				IP:   localP.IP,
				Port: localP.Port,
			})
		}
		return
	case []interface{}:
		//FIXME: vars.Add("http responses with list peers", 1)
		for _, i := range v {
			var localP Peer
			localP.fromDictInterface(i.(map[string]interface{}))
			*me = append(*me, tracker.Peer{
				IP:   localP.IP,
				Port: localP.Port,
			})
		}
		return
	default:
		err = fmt.Errorf("unsupported type: %T", _v)
		return
	}
}

func setupQuery(url *url.URL, ar tracker.AnnounceRequest) {
	rawQuery := url.RawQuery
	/*rawQuery = strings.Replace(rawQuery, "{infohash}", string(ar.InfoHash[:]), 1)
	rawQuery = strings.Replace(rawQuery, "{peerid}", peerid.PeerId(ar.PeerId).Format(), 1)
	rawQuery = strings.Replace(rawQuery, "{key}", ar.Key, 1)
	rawQuery = strings.Replace(rawQuery, "{port}", fmt.Sprintf("%d", ar.Port), 1)
	rawQuery = strings.Replace(rawQuery, "{uploaded}", strconv.FormatInt(ar.Uploaded, 10), 1)
	rawQuery = strings.Replace(rawQuery, "{downloaded}", strconv.FormatInt(ar.Downloaded, 10), 1)
	rawQuery = strings.Replace(rawQuery, "{left}", strconv.FormatInt(ar.Left, 10), 1)*/

	/*
		q.Set("info_hash", string(ar.InfoHash[:]))
		q.Set("peer_id", string(ar.PeerId[:]))
		q.Set("port", fmt.Sprintf("%d", ar.Port))
		q.Set("uploaded", strconv.FormatInt(ar.Uploaded, 10))
		q.Set("downloaded", strconv.FormatInt(ar.Downloaded, 10))
		q.Set("left", strconv.FormatInt(ar.Left, 10))
		if ar.Event != tracker.None {
			q.Set("event", ar.Event.String())
		}
		q.Set("compact", "1")
		q.Set("supportcrypto", "1")*/
	/*if opts.ClientIp4.IP != nil {
		q.Set("ipv4", opts.ClientIp4.String())
	}
	if opts.ClientIp6.IP != nil {
		q.Set("ipv6", opts.ClientIp6.String())
	}*/
	url.RawQuery = rawQuery
}
