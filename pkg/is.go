package is

import (
	"log"
)

// Wayback is the handle of saving webpages to archive.is
func (wbrc *Archiver) Wayback(links []string) (map[string]string, error) {
	collect := make(map[string]string)
	for _, link := range links {
		if !isURL(link) {
			log.Print(link + " is invalid url.")
			continue
		}
		collect[link] = link
	}

	ch := make(chan string, len(collect))
	defer close(ch)

	for link := range collect {
		go wbrc.fetch(link, ch)
		collect[link] = <-ch
	}

	return collect, nil
}
