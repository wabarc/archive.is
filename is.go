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
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/wabarc/helper"
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
	onion     = "archivecaslytosk.onion" // archiveiya74codqgiixo33q62qlrqtkgmcitqx5u2oeqnmn5bpcbiyd.onion
	cookie    = ""
	timeout   = 120 * time.Second
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
	if os.Getenv("DEBUG") != "" {
		logger.EnableDebug()
	}
}

// Wayback is the handle of saving webpages to archive.is
func (wbrc *Archiver) Wayback(links []string) (map[string]string, error) {
	collects, results := make(map[string]string), make(map[string]string)
	for _, link := range links {
		if helper.IsURL(link) {
			collects[link] = link
		}
	}
	if len(collects) == 0 {
		return results, fmt.Errorf("Not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	torClient, t, err := newTorClient(ctx)
	defer closeTor(t)
	if err != nil {
		logger.Error("%v", err)
	}

	is := &IS{
		wbrc:       wbrc,
		httpClient: &http.Client{Timeout: timeout, CheckRedirect: noRedirect},
		torClient:  torClient,
	}

	ch := make(chan string, len(collects))
	defer close(ch)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, link := range collects {
		wg.Add(1)
		go func(link string) {
			mu.Lock()
			is.submitid = ""
			is.archive(link, ch)
			results[link] = strings.Replace(<-ch, onion, "archive.today", 1)
			mu.Unlock()
			wg.Done()
		}(link)
	}
	wg.Wait()

	if len(results) == 0 {
		return results, fmt.Errorf("No results")
	}

	return results, nil
}

// Playback handle searching archived webpages from archive.is
func (wbrc *Archiver) Playback(links []string) (map[string]string, error) {
	collects, results := make(map[string]string), make(map[string]string)
	for _, link := range links {
		if helper.IsURL(link) {
			collects[link] = link
		}
	}
	if len(collects) == 0 {
		return results, fmt.Errorf("Not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	torClient, t, err := newTorClient(ctx)
	defer closeTor(t)
	if err != nil {
		logger.Error("%v", err)
	}

	is := &IS{
		wbrc:       wbrc,
		httpClient: &http.Client{Timeout: timeout, CheckRedirect: noRedirect},
		torClient:  torClient,
	}

	ch := make(chan string, len(collects))
	defer close(ch)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, link := range collects {
		wg.Add(1)
		go func(link string) {
			mu.Lock()
			is.submitid = ""
			is.search(link, ch)
			results[link] = strings.Replace(<-ch, onion, "archive.today", 1)
			mu.Unlock()
			wg.Done()
		}(link)
	}
	wg.Wait()

	if len(results) == 0 {
		return results, fmt.Errorf("No results")
	}

	return results, nil
}
func (is *IS) archive(uri string, ch chan<- string) {
	endpoint, err := is.getValidDomain()
	if err != nil {
		ch <- fmt.Sprint("archive.today is unavailable.")
		return
	}

	if is.wbrc.Anyway != "" {
		anyway = is.wbrc.Anyway
	}
	data := url.Values{
		"submitid": {is.submitid},
		"anyway":   {anyway},
		"url":      {uri},
	}
	domain := endpoint.String()
	req, err := http.NewRequest("POST", domain+"/submit/", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", domain)
	req.Header.Add("Origin", domain)
	req.Header.Add("Host", endpoint.Hostname())
	req.Header.Add("Cookie", is.getCookie())
	resp, err := is.httpClient.Do(req)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode / 100
	if code == 1 || code == 4 || code == 5 {
		final := fmt.Sprintf("%s?url=%s", domain, uri)
		ch <- final
		return
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}

	// When use anyway parameter.
	refresh := resp.Header.Get("Refresh")
	if len(refresh) > 0 {
		r := strings.Split(refresh, ";url=")
		if len(r) == 2 {
			ch <- r[1]
			return
		}
	}
	loc := resp.Header.Get("location")
	if len(loc) > 2 {
		ch <- loc
		return
	}
	// Redirect to final url if page saved.
	final := resp.Request.URL.String()
	if len(final) > 0 && strings.Contains(final, "/submit/") == false {
		ch <- final
		return
	}

	ch <- fmt.Sprintf("%s/timegate/%s", domain, uri)
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
	if strings.Contains(url, "http") == false {
		return "", fmt.Errorf("missing protocol scheme")
	}

	r := strings.NewReader("")
	req, err := http.NewRequest("GET", url, r)
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

func (is *IS) search(uri string, ch chan<- string) {
	endpoint, err := is.getValidDomain()
	if err != nil {
		ch <- fmt.Sprint("archive.today is unavailable.")
		return
	}

	domain := endpoint.String()
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", domain, uri), nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", domain)
	req.Header.Add("Host", endpoint.Hostname())
	resp, err := is.httpClient.Do(req)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}

	target, exists := doc.Find("#row0 > .TEXT-BLOCK > a").Attr("href")
	if !exists {
		ch <- "Not found"
		return
	}

	ch <- target
}
