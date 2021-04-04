package is

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/wabarc/helper"
)

type Archiver struct {
	Anyway string
	Cookie string

	DialContext         func(ctx context.Context, network, addr string) (net.Conn, error)
	SkipTLSVerification bool
}

// Wayback is the handle of saving webpages to archive.is
func (wbrc *Archiver) Wayback(links []string) (map[string]string, error) {
	collect, results := make(map[string]string), make(map[string]string)
	for _, link := range links {
		if !helper.IsURL(link) {
			log.Print(link + " is invalid url.")
			continue
		}
		collect[link] = link
	}

	torClient, tor, err := newTorClient()
	if err != nil {
		log.Println(err)
	} else {
		defer tor.Close()
	}

	ch := make(chan string, len(collect))
	defer close(ch)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for link := range collect {
		wg.Add(1)
		go func(link string) {
			is := &IS{
				wbrc:       wbrc,
				httpClient: &http.Client{Timeout: timeout},
				torClient:  torClient,
				final:      "",
				submitid:   "",
			}
			is.fetch(link, ch)
			mu.Lock()
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
