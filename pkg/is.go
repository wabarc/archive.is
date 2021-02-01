package is

import (
	"log"
	"strings"
	"sync"

	"github.com/wabarc/helper"
)

// Wayback is the handle of saving webpages to archive.is
func (wbrc *Archiver) Wayback(links []string) (map[string]string, error) {
	collect := make(map[string]string)
	for _, link := range links {
		if !helper.IsURL(link) {
			log.Print(link + " is invalid url.")
			continue
		}
		collect[link] = link
	}

	if client, tor, err := wbrc.newTorClient(); err != nil {
		log.Println(err)
	} else {
		wbrc.torClient = client
		defer tor.Close()
	}

	ch := make(chan string, len(collect))
	defer close(ch)

	var wg sync.WaitGroup
	for link := range collect {
		wg.Add(1)
		go func(link string, ch chan string) {
			wbrc.fetch(link, ch)
			collect[link] = strings.Replace(<-ch, onion, "archive.today", 1)
			wg.Done()
		}(link, ch)
	}
	wg.Wait()

	return collect, nil
}
