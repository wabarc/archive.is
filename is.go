package is

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/wabarc/logger"
)

type Archiver struct {
	Anyway string
	Cookie string

	DialContext         func(ctx context.Context, network, addr string) (net.Conn, error)
	SkipTLSVerification bool
}

type IS struct {
	wbrc *Archiver

	submitid string

	httpClient *http.Client
	torClient  *http.Client
}

var (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36"
	anyway    = "0"
	scheme    = "http"
	onion     = "archiveiya74codqgiixo33q62qlrqtkgmcitqx5u2oeqnmn5bpcbiyd.onion" // archivecaslytosk.onion
	cookie    = ""
	domains   = []string{
		"archive.today",
		"archive.is",
		"archive.li",
		"archive.vn",
		"archive.fo",
		"archive.md",
		"archive.ph",
	}
)

func init() {
	debug := os.Getenv("DEBUG")
	if debug == "true" || debug == "1" || debug == "on" {
		logger.EnableDebug()
	}
}

// Wayback is the handle of saving webpages to archive.is
func (wbrc *Archiver) Wayback(ctx context.Context, in *url.URL) (dst string, err error) {
	torClient, t, err := newTorClient(ctx)
	defer closeTor(t) // nolint:errcheck
	if err != nil {
		logger.Error("%v", err)
	}

	is := &IS{
		wbrc:       wbrc,
		httpClient: &http.Client{CheckRedirect: noRedirect},
		torClient:  torClient,
	}

	dst, err = is.archive(ctx, in)
	if err != nil {
		return
	}
	dst = strings.Replace(dst, onion, "archive.today", 1)

	return
}

// Playback handle searching archived webpages from archive.is
func (wbrc *Archiver) Playback(ctx context.Context, in *url.URL) (dst string, err error) {
	torClient, t, err := newTorClient(ctx)
	defer closeTor(t) // nolint:errcheck
	if err != nil {
		logger.Error("%v", err)
	}

	is := &IS{
		wbrc:       wbrc,
		httpClient: &http.Client{CheckRedirect: noRedirect},
		torClient:  torClient,
	}

	dst, err = is.search(ctx, in)
	if err != nil {
		return
	}
	dst = strings.Replace(dst, onion, "archive.today", 1)

	return
}
func (is *IS) archive(ctx context.Context, u *url.URL) (string, error) {
	endpoint, err := is.getValidDomain()
	if err != nil {
		return "", fmt.Errorf("archive.today is unavailable.")
	}

	if is.wbrc.Anyway != "" {
		anyway = is.wbrc.Anyway
	}
	uri := u.String()
	data := url.Values{
		"submitid": {is.submitid},
		"anyway":   {anyway},
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
	req.Header.Add("Cookie", is.getCookie())
	resp, err := is.httpClient.Do(req)
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

func (is *IS) getCookie() string {
	c := os.Getenv("ARCHIVE_COOKIE")
	if c != "" {
		is.wbrc.Cookie = c
	}

	if is.wbrc.Cookie != "" {
		return is.wbrc.Cookie
	} else {
		return cookie
	}
}

func (is *IS) getSubmitID(url string) (string, error) {
	if !strings.Contains(url, "http") {
		return "", fmt.Errorf("missing protocol scheme")
	}

	r := strings.NewReader("")
	req, _ := http.NewRequest("GET", url, r)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Cookie", is.getCookie())
	resp, err := is.httpClient.Do(req)
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

func (is *IS) getValidDomain() (*url.URL, error) {
	var endpoint *url.URL
	// get valid domain and submitid
	r := func(domains []string) {
		for _, domain := range domains {
			h := fmt.Sprintf("%v://%v", scheme, domain)
			id, err := is.getSubmitID(h)
			if err != nil {
				continue
			}
			is.submitid = id
			endpoint, _ = url.Parse(h)
			break
		}
	}

	// Try request over Tor hidden service.
	if is.torClient != nil {
		is.httpClient = is.torClient

		r([]string{onion})
	}

	if endpoint == nil || is.submitid == "" {
		r(domains)
		if endpoint == nil || is.submitid == "" {
			return nil, fmt.Errorf("archive.today is unavailable.")
		}
	}

	return endpoint, nil
}

func (is *IS) search(ctx context.Context, in *url.URL) (string, error) {
	endpoint, err := is.getValidDomain()
	if err != nil {
		return "", fmt.Errorf("archive.today is unavailable.")
	}

	uri := in.String()
	domain := endpoint.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", domain, uri), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", domain)
	req.Header.Add("Host", endpoint.Hostname())
	resp, err := is.httpClient.Do(req)
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
