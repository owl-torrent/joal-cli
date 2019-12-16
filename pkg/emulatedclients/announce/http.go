package announce

import (
	"bytes"
	"fmt"
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

type IHttpAnnouncer interface {
	Announce(url url.URL, announceRequest AnnounceRequest) (tracker.AnnounceResponse, error)
	AfterPropertiesSet() error
}

type HttpAnnouncer struct {
	UrlEncoder     urlencoder.UrlEncoder `yaml:"urlEncoder"`
	Query          string                `yaml:"query" validate:"required"`
	RequestHeaders []HttpRequestHeader   `yaml:"requestHeaders" validate:"dive"`
	queryTemplate  *template.Template    `yaml:"-"`
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
	if len(_url.Query()) > 0 {
		queryString = fmt.Sprintf("%s&%s", url.RawQuery, queryString)
	}
	_url.RawQuery = queryString

	req, err := http.NewRequest("GET", _url.String(), nil)
	if err != nil {
		return
	}

	for _, v := range a.RequestHeaders {
		req.Header.Add(v.Name, v.Value)
	}

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
	var trackerResponse tracker.HttpResponse
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
	Name  string `yaml:"name" validate:"required"`
	Value string `yaml:"value" validate:"required"`
}
