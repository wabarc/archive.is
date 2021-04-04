package is

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type IS struct {
	wbrc *Archiver

	submitid string
	final    string

	baseuri    *url.URL
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

func (is *IS) fetch(s string, ch chan<- string) {
	// get valid domain and submitid
	r := func(domains []string) {
		for _, domain := range domains {
			h := fmt.Sprintf("%v://%v", scheme, domain)
			id, err := is.getSubmitID(h)
			if err != nil {
				continue
			}
			is.baseuri, _ = url.Parse(h)
			is.submitid = id
			break
		}
	}

	// Try request over Tor hidden service.
	if is.torClient != nil {
		is.httpClient = is.torClient

		r([]string{onion})
	}

	if is.baseuri == nil || is.submitid == "" {
		r(domains)
		if is.baseuri == nil || is.submitid == "" {
			ch <- fmt.Sprint("archive.today is unavailable.")
			return
		}
	}

	if is.wbrc.Anyway != "" {
		anyway = is.wbrc.Anyway
	}
	data := url.Values{
		"submitid": {is.submitid},
		"anyway":   {anyway},
		"url":      {s},
	}
	uri := is.baseuri.String()
	req, err := http.NewRequest("POST", is.baseuri.String()+"/submit/", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", uri)
	req.Header.Add("Origin", uri)
	req.Header.Add("Host", is.baseuri.Hostname())
	req.Header.Add("Cookie", is.getCookie())
	resp, err := is.httpClient.Do(req)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode / 100
	if code == 1 || code == 4 || code == 5 {
		final := fmt.Sprintf("%s?url=%s", uri, s)
		ch <- final
		return
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}

	// Redirect to final url if page saved.
	final := resp.Request.URL.String()
	if len(final) > 0 && strings.Contains(final, "/submit/") == false {
		is.final = final
	}
	loc := resp.Header.Get("location")
	if len(loc) > 2 {
		is.final = loc
	}
	// When use anyway parameter.
	refresh := resp.Header.Get("refresh")
	if len(refresh) > 0 {
		r := strings.Split(refresh, ";url=")
		if len(r) == 2 {
			is.final = r[1]
		}
	}

	ch <- is.final
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
