package announcer

import (
	"compress/gzip"
	"context"
	"fmt"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/urlencoder"
	"github.com/anthonyraymond/joal-cli/internal/old/core/logs"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
)

type IHttpAnnouncer interface {
	iAnnouncer
	AfterPropertiesSet(proxyFunc func(*http.Request) (*url.URL, error)) error
}

type HttpAnnouncer struct {
	UrlEncoder     urlencoder.UrlEncoder `yaml:"urlEncoder"`
	Query          string                `yaml:"query" validate:"required"`
	RequestHeaders []HttpRequestHeader   `yaml:"requestHeaders" validate:"dive"`
	queryTemplate  *template.Template    `yaml:"-"`
	httpClient     *http.Client          `yaml:"-"`
}

func (a *HttpAnnouncer) AfterPropertiesSet(proxyFunc func(*http.Request) (*url.URL, error)) error {
	var err error

	a.queryTemplate, err = template.New("httpQueryTemplate").Funcs(TemplateFunctions(&a.UrlEncoder)).Parse(a.Query)
	if err != nil {
		return err
	}

	a.httpClient = &http.Client{
		Timeout: time.Second * 15,
		Transport: &http.Transport{
			Proxy:              proxyFunc,
			DisableCompression: true, // Disable auto send of Accept-Encoding gzip header. Since the lib dont add the header on it's own we'll have to handle the gzip decompression on our own
			DialContext: (&net.Dialer{
				Timeout: 15 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 15 * time.Second,
			//DisableKeepAlives:   true, // see https://github.com/anacrolix/torrent/commit/04ff050ecd5f5beab9b20a0f4170fda1e71062a4
		},
	}
	return nil
}

func (a *HttpAnnouncer) Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error) {
	log := logs.GetLogger()
	_url := copyURL(&url)
	queryString, err := buildQueryString(a.queryTemplate, announceRequest)
	if err != nil {
		return AnnounceResponse{}, fmt.Errorf("fail to format query string: %w", err)
	}
	if len(_url.Query()) > 0 {
		queryString = fmt.Sprintf("%s&%s", url.RawQuery, queryString)
	}
	_url.RawQuery = queryString

	req, err := http.NewRequestWithContext(ctx, "GET", _url.String(), nil)
	if err != nil {
		return AnnounceResponse{}, err
	}

	for _, v := range a.RequestHeaders {
		req.Header.Add(v.Name, v.Value)
	}
	log.Debug("announce details",
		zap.String("protocol", req.Proto),
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Reflect("headers", req.Header),
		zap.ByteString("infohash", announceRequest.InfoHash[:]),
	)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return AnnounceResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	bodyBytes, err := readResponseBody(resp)
	if err != nil {
		return AnnounceResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != 200 {
		return AnnounceResponse{}, fmt.Errorf("response from tracker: %s: %s", resp.Status, fmt.Sprintf("%x", bodyBytes))
	}
	var trackerResponse tracker.HttpResponse
	err = bencode.Unmarshal(bodyBytes, &trackerResponse)
	if _, ok := err.(bencode.ErrUnusedTrailingBytes); !ok && err != nil {
		return AnnounceResponse{}, fmt.Errorf("error decoding %q: %w", bodyBytes, err)
	}
	if trackerResponse.FailureReason != "" {
		return AnnounceResponse{}, fmt.Errorf("tracker gave failure reason: %q", trackerResponse.FailureReason)
	}
	ret := AnnounceResponse{
		Interval: time.Duration(trackerResponse.Interval) * time.Second,
		Leechers: trackerResponse.Incomplete,
		Seeders:  trackerResponse.Complete,
		Peers:    trackerResponse.Peers,
	}
	for _, na := range trackerResponse.Peers6 {
		ret.Peers = append(ret.Peers, tracker.Peer{
			IP:   na.IP,
			Port: na.Port,
		})
	}
	return ret, nil
}

func buildQueryString(queryTemplate *template.Template, ar AnnounceRequest) (string, error) {
	sb := strings.Builder{}
	err := queryTemplate.Execute(&sb, ar)
	return sb.String(), err
}

func readResponseBody(response *http.Response) ([]byte, error) {
	var reader = response.Body

	if response.Header.Get("Content-Encoding") == "gzip" {
		var err error
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to decode gzip body content: %w", err)
		}
		defer func() { _ = reader.Close() }()
	}

	return ioutil.ReadAll(reader)
}

type HttpRequestHeader struct {
	Name  string `yaml:"name" validate:"required"`
	Value string `yaml:"value" validate:"required"`
}

func copyURL(u *url.URL) (ret *url.URL) {
	ret = new(url.URL)
	*ret = *u
	if u.User != nil {
		ret.User = new(url.Userinfo)
		*ret.User = *u.User
	}
	return
}
