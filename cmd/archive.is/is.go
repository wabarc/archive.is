package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/wabarc/archive.is"
)

func main() {
	var (
		playback bool
		version  bool
	)

	const playbackHelp = "Search archived URL"
	const versionHelp = "Show version"

	flag.BoolVar(&playback, "playback", false, playbackHelp)
	flag.BoolVar(&playback, "p", false, playbackHelp)
	flag.BoolVar(&version, "version", false, versionHelp)
	flag.BoolVar(&version, "v", false, versionHelp)
	flag.Parse()

	if version {
		fmt.Println(is.Version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		e := os.Args[0]
		fmt.Printf("  %s url [url]\n\n", e)
		fmt.Printf("example:\n  %s https://example.com https://example.org\n\n", e)
		os.Exit(1)
	}

	arc := is.NewArchiver(nil)
	defer arc.CloseTor() // nolint:errcheck

	if playback {
		process(arc.Playback, args)
		os.Exit(0)
	}

	process(arc.Wayback, args)
}

func process(f func(context.Context, *url.URL) (string, error), args []string) {
	var wg sync.WaitGroup
	for _, arg := range args {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			u, err := url.Parse(link)
			if err != nil {
				fmt.Println(link, "=>", fmt.Sprintf("%v", err))
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			r, err := f(ctx, u)
			if err != nil {
				fmt.Println(link, "=>", fmt.Sprintf("%v", err))
				return
			}
			fmt.Println(link, "=>", r)
		}(arg)
	}
	wg.Wait()
}
