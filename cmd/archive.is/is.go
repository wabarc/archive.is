package main

import (
	"flag"
	"fmt"
	"os"

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

	wbrc := &is.Archiver{}

	if playback {
		collects, _ := wbrc.Playback(args)
		for orig, dest := range collects {
			fmt.Println(orig, "=>", dest)
		}
		os.Exit(0)
	}

	saved, _ := wbrc.Wayback(args)
	for orig, dest := range saved {
		fmt.Println(orig, "=>", dest)
	}
}
