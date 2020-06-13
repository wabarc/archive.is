package is

import (
	"fmt"
	"time"
)

func Wayback(links []string) []string {
	start := time.Now()
	worklist := make(map[string]string)
	for _, link := range links {
		if !isURL(link) {
			fmt.Println(link + " is invalid url.")
			continue
		}
		worklist[link] = link
	}

	var collect []string
	if len(worklist) < 1 {
		return collect
	}

	ch := make(chan string, len(worklist))
	defer close(ch)

	for link := range worklist {
		go fetch(link, ch)
		collect = append(collect, <-ch)
	}

	fmt.Printf("%.2fs elapsed\n\n", time.Since(start).Seconds())

	return collect
}
