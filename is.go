package is

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cretz/bine/tor"
	"github.com/wabarc/logger"
)

const timeout = 30 * time.Second

var (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; rv:102.0) Gecko/20100101 Firefox/102.0"
	onion     = "http://archiveiya74codqgiixo33q62qlrqtkgmcitqx5u2oeqnmn5bpcbiyd.onion" // archivecaslytosk.onion
	domains   = []string{
		"https://archive.ph",
		"https://archive.today",
		"https://archive.is",
		"https://archive.li",
		"https://archive.vn",
		"https://archive.fo",
		"https://archive.md",
	}
)

// Archiver represents an archiver that can be used to submit and search for
// archived versions of web pages on archive.today.
type Archiver struct {
	*http.Client

	torClient *http.Client
	tor       *tor.Tor

	// Cookie string for setting cookies in requests.
	Cookie string

	anyway string
}

type IS struct {
	submitid string
}

func init() {
	debug := os.Getenv("DEBUG")
	if debug == "true" || debug == "1" || debug == "on" {
		logger.EnableDebug()
	}
}

// NewArchiver returns a Archiver struct with the specified HTTP client and Tor network client.
// It's the responsibility of the caller to call CloseTor when it is no longer needed.
func NewArchiver(client *http.Client) *Archiver {
	arc := &Archiver{anyway: "1"}
	if client == nil {
		client = &http.Client{}
	}
	client.Timeout = timeout
	// client.CheckRedirect = noRedirect
	arc.Client = client

	torClient, tor, err := newTorClient(context.Background())
	if err != nil {
		return arc
	}
	if torClient != nil {
		arc.torClient = torClient
	}
	arc.tor = tor

	return arc
}

// CloseTor closes the Tor client if it exists.
func (arc *Archiver) CloseTor() {
	if arc.tor != nil {
		_ = closeTor(arc.tor)
	}
}

// Do implements the http.Do method to execute an HTTP request using the embedded HTTP
// client or the Tor client as a fallback if the primary client fails to execute the request.
func (arc *Archiver) Do(req *http.Request) (resp *http.Response, err error) {
	if strings.HasSuffix(req.URL.Hostname(), ".onion") {
		goto tryTor
	}

	resp, err = arc.Client.Do(req)
	if err != nil {
		goto tryTor
	}
	return

tryTor:
	if arc.torClient != nil {
		return arc.torClient.Do(req)
	}
	return
}

// Wayback is the handle of saving webpages to archive.is
func (arc *Archiver) Wayback(ctx context.Context, in *url.URL) (dst string, err error) {
	is := &IS{}
	dst, err = arc.archive(ctx, is, in)
	if err != nil {
		return
	}
	dst = strings.Replace(dst, onion, "https://archive.today", 1)
	dst = regexp.MustCompile(`\/again\?url=.*`).ReplaceAllString(dst, "")

	return
}

// Playback handle searching archived webpages from archive.is
func (arc *Archiver) Playback(ctx context.Context, in *url.URL) (dst string, err error) {
	is := &IS{}
	dst, err = arc.search(ctx, is, in)
	if err != nil {
		return
	}
	dst = strings.Replace(dst, onion, "https://archive.today", 1)

	return
}

func (arc *Archiver) archive(ctx context.Context, is *IS, u *url.URL) (string, error) {
	endpoint, err := arc.getValidDomain(is)
	if err != nil {
		return "", fmt.Errorf("archive.today is unavailable.")
	}

	uri := u.String()
	data := url.Values{
		"submitid": {is.submitid},
		"anyway":   {arc.anyway},
		"url":      {uri},
	}
	domain := endpoint.String()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, domain+"/submit/", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", domain)
	req.Header.Add("Origin", domain)
	req.Header.Add("Host", endpoint.Hostname())
	req.Header.Add("Cookie", getCookie())
	resp, err := arc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	code := resp.StatusCode / 100
	if code == 1 || code == 4 || code == 5 {
		final := fmt.Sprintf("%s?url=%s", domain, uri)
		return final, nil
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return "", err
	}

	// When use anyway parameter.
	refresh := resp.Header.Get("Refresh")
	if len(refresh) > 0 {
		r := strings.Split(refresh, ";url=")
		if len(r) == 2 {
			return r[1], nil
		}
	}
	loc := resp.Header.Get("location")
	if len(loc) > 2 {
		return loc, nil
	}
	// Redirect to final url if page saved.
	final := resp.Request.URL.String()
	if len(final) > 0 && !strings.Contains(final, "/submit/") {
		return final, nil
	}

	return fmt.Sprintf("%s/timegate/%s", domain, uri), nil
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func getCookie() string {
	return os.Getenv("ARCHIVE_COOKIE")
}

func (arc *Archiver) getSubmitID(url string) (string, error) {
	r := strings.NewReader("")
	req, _ := http.NewRequest("GET", url, r)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Cookie", getCookie())
	resp, err := arc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code error: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	id, exists := doc.Find("input[name=submitid]").First().Attr("value")
	if !exists {
		return "", fmt.Errorf("submitid not found")
	}

	return id, nil
}

func (arc *Archiver) getValidDomain(is *IS) (*url.URL, error) {
	var endpoint *url.URL
	// get valid domain and submitid
	try := func(domains ...string) {
		for _, domain := range domains {
			id, err := arc.getSubmitID(domain)
			if err != nil {
				continue
			}
			is.submitid = id
			endpoint, _ = url.Parse(domain)
			break
		}
	}

	if endpoint == nil || is.submitid == "" {
		try(domains...)
		if endpoint != nil && is.submitid != "" {
			return endpoint, nil
		}
	}

	// Try request over Tor hidden service.
	if arc.torClient != nil {
		try(onion)
	}
	if endpoint == nil {
		return nil, fmt.Errorf("not found valid domain")
	}

	return endpoint, nil
}

func (arc *Archiver) search(ctx context.Context, is *IS, in *url.URL) (string, error) {
	endpoint, err := arc.getValidDomain(is)
	if err != nil {
		return "", fmt.Errorf("archive.today is unavailable.")
	}

	domain := endpoint.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", domain, in), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", domain)
	req.Header.Add("Host", endpoint.Hostname())
	resp, err := arc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	target, exists := doc.Find("#row0 > .TEXT-BLOCK > a").Attr("href")
	if !exists {
		return "", fmt.Errorf("Not found")
	}

	return target, nil
}
