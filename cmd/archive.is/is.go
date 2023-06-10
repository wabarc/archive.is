package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/wabarc/archive.is"
	"github.com/wabarc/proxier"
)

func main() {
	var (
		proxy string

		playback bool
		version  bool
	)

	const proxyHelp = "Proxy server, e.g. socks5://127.0.0.1:1080"
	const playbackHelp = "Search archived URL"
	const versionHelp = "Show version"

	flag.StringVar(&proxy, "proxy", "", proxyHelp)
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

	client := &http.Client{}
	if proxy != "" {
		rt, err := proxier.NewUTLSRoundTripper(proxier.Proxy(proxy))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		client.Transport = rt
	}

	arc := is.NewArchiver(client)
	defer arc.CloseTor()

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
