package announce

import (
	"bytes"
	"fmt"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/missinggo/httptoo"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/urlencoder"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
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
	Announce(url url.URL, announceRequest AnnounceRequest) (tracker.AnnounceResponse, error)
	AfterPropertiesSet() error
}

type HttpAnnouncer struct {
	UrlEncoder      urlencoder.UrlEncoder `yaml:"urlEncoder"`
	ShouldUrlEncode bool                  `yaml:"shouldUrlEncode"`
	Query           string                `yaml:"string"`
	RequestHeaders  []HttpRequestHeader   `yaml:"requestHeaders"`
	queryTemplate   *template.Template    `yaml:"-"`
}

func (a *HttpAnnouncer) AfterPropertiesSet() error {
	var err error
	a.queryTemplate, err = template.New("httpQueryTemplate").Funcs(TemplateFunctions(&a.UrlEncoder)).Parse(a.Query)
	if err != nil {
		return err
	}
	return nil
}

func (a *HttpAnnouncer) Announce(url url.URL, announceRequest AnnounceRequest) (ret tracker.AnnounceResponse, err error) {
	_url := httptoo.CopyURL(&url)
	queryString, err := buildQueryString(a.queryTemplate, announceRequest)
	if err != nil {
		return ret, errors.Wrap(err, "fail to format query string")
	}
	if len(_url.Query()) >= 0 {
		queryString = fmt.Sprintf("&%s", queryString)
	}
	_url.RawQuery = queryString

	req, err := http.NewRequest("GET", _url.String(), nil)

	resp, err := (&http.Client{
		Timeout: time.Second * 15,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 15 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 15 * time.Second,
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

func buildQueryString(queryTemplate *template.Template, ar AnnounceRequest) (string, error) {
	sb := strings.Builder{}
	err := queryTemplate.Execute(&sb, ar)
	return sb.String(), err
}

type HttpRequestHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Peers []tracker.Peer

func (p *Peers) UnmarshalBencode(b []byte) (err error) {
	var _v interface{}
	err = bencode.Unmarshal(b, &_v)
	if err != nil {
		return
	}
	switch v := _v.(type) {
	case string:
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
			*p = append(*p, tracker.Peer{
				IP:   localP.IP,
				Port: localP.Port,
			})
		}
		return
	case []interface{}:
		for _, i := range v {
			var localP Peer
			localP.FromDictInterface(i.(map[string]interface{}))
			*p = append(*p, tracker.Peer{
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
