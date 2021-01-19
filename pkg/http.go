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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Archiver struct {
	Anyway string
	Cookie string

	DialContext         func(ctx context.Context, network, addr string) (net.Conn, error)
	SkipTLSVerification bool

	final    string
	submitid string
	isTor    bool

	httpClient *http.Client
	torClient  *http.Client
}

var (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36"
	anyway    = "0"
	scheme    = "http"
	onion     = "archivecaslytosk.onion" // archiveiya74codqgiixo33q62qlrqtkgmcitqx5u2oeqnmn5bpcbiyd.onion
	cookie    = "cf_clearance=dd7e157eb2d43acf2decfafd13c650dd80d825b5-1600696752-KXZXFYWE"
	timeout   = 120 * time.Second
	baseuri   *url.URL
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

func (wbrc *Archiver) fetch(s string, ch chan<- string) {
	wbrc.httpClient = &http.Client{
		Timeout: timeout,
	}

	// get valid domain and submitid
	r := func(domains []string) {
		for _, domain := range domains {
			h := fmt.Sprintf("%v://%v", scheme, domain)
			id, err := wbrc.getSubmitID(h)
			if err != nil {
				continue
			}
			baseuri, _ = url.Parse(h)
			wbrc.submitid = id
			break
		}
	}
	r(domains)

	if baseuri == nil || wbrc.submitid == "" {
		// Try request over Tor hidden service.
		if wbrc.torClient == nil {
			ch <- fmt.Sprint("Tor network unreachable.")
			return
		}
		wbrc.httpClient = wbrc.torClient

		r([]string{onion})
		if baseuri == nil || wbrc.submitid == "" {
			ch <- fmt.Sprint("archive.today is unavailable.")
			return
		}
	}

	if wbrc.Anyway != "" {
		anyway = wbrc.Anyway
	}
	data := url.Values{
		"submitid": {wbrc.submitid},
		"anyway":   {anyway},
		"url":      {s},
	}
	req, err := http.NewRequest("POST", baseuri.String()+"/submit/", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", baseuri.String())
	req.Header.Add("Origin", baseuri.String())
	req.Header.Add("Host", baseuri.Hostname())
	req.Header.Add("Cookie", wbrc.getCookie())
	resp, err := wbrc.httpClient.Do(req)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode / 100
	if code == 1 || code == 4 || code == 5 {
		final := fmt.Sprintf("%v?url=%s", baseuri.String(), s)
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
		wbrc.final = final
	}
	loc := resp.Header.Get("location")
	if len(loc) > 2 {
		wbrc.final = loc
	}
	// When use anyway parameter.
	refresh := resp.Header.Get("refresh")
	if len(refresh) > 0 {
		r := strings.Split(refresh, ";url=")
		if len(r) == 2 {
			wbrc.final = r[1]
		}
	}

	ch <- wbrc.final
}

func (wbrc *Archiver) getCookie() string {
	c := os.Getenv("ARCHIVE_COOKIE")
	if c != "" {
		wbrc.Cookie = c
	}

	if wbrc.Cookie != "" {
		return wbrc.Cookie
	} else {
		return cookie
	}
}

func (wbrc *Archiver) getSubmitID(url string) (string, error) {
	if strings.Contains(url, "http") == false {
		return "", fmt.Errorf("missing protocol scheme")
	}

	r := strings.NewReader("")
	req, err := http.NewRequest("GET", url, r)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Cookie", wbrc.getCookie())
	resp, err := wbrc.httpClient.Do(req)
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

func isURL(str string) bool {
	re := regexp.MustCompile(`https?://?[-a-zA-Z0-9@:%._\+~#=]{1,255}\.[a-z]{0,63}\b(?:[-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
	match := re.FindAllString(str, -1)

	return len(match) >= 1
}
