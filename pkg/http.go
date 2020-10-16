package is

import (
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
)

type Archiver struct {
	Anyway string
	Cookie string

	url      string
	final    string
	submitid string
}

var (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36"
	anyway    = "0"
	scheme    = "https"
	cookie    = "cf_clearance=dd7e157eb2d43acf2decfafd13c650dd80d825b5-1600696752-KXZXFYWE"
	timeout   = time.Duration(30) * time.Second
	domains   = []string{
		"archive.li",
		"archive.vn",
		"archive.fo",
		"archive.md",
		"archive.ph",
		"archive.today",
		"archive.is",
	}
)

func (wbrc *Archiver) fetch(s string, ch chan<- string) {
	// get valid domain and submitid
	for _, domain := range domains {
		h := fmt.Sprintf("%v://%v", scheme, domain)
		id, err := wbrc.getSubmitID(h)
		if err != nil {
			continue
		}
		wbrc.url = h + "/submit/"
		wbrc.submitid = id
		break
	}

	if len(wbrc.url) < 1 || len(wbrc.submitid) < 1 {
		ch <- fmt.Sprint("Archive.today is unavailable.")
		return
	}

	if wbrc.Anyway != "" {
		anyway = wbrc.Anyway
	}
	client := &http.Client{
		Timeout: timeout,
	}
	data := url.Values{
		"submitid": {wbrc.submitid},
		"anyway":   {anyway},
		"url":      {s},
	}
	req, err := http.NewRequest("POST", wbrc.url, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Cookie", wbrc.getCookie())
	resp, err := client.Do(req)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()

	code := resp.StatusCode / 100
	if code == 1 || code == 4 || code == 5 {
		ch <- fmt.Sprint("no access")
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

	client := &http.Client{
		Timeout: timeout,
	}

	r := strings.NewReader("")
	req, err := http.NewRequest("GET", url, r)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Cookie", wbrc.getCookie())
	resp, err := client.Do(req)

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
	re := regexp.MustCompile(`(http(s)?:\/\/.)?(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
	match := re.FindAllString(str, -1)
	for _, el := range match {
		if len(el) > 2 {
			return true
		}
	}
	return false
}
