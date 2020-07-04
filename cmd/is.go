package is

import (
	"flag"
	"fmt"
	"os"

	"github.com/wabarc/archive.is/pkg"
)

func Run() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		e := os.Args[0]
		fmt.Printf("  %s url [url]\n\n", e)
		fmt.Printf("example:\n  %s https://www.google.com https://www.bbc.co.uk/\n\n", e)
		os.Exit(1)
	}

	wbrc := &is.Archiver{}
	saved, _ := wbrc.Wayback(args)
	for orig, dest := range saved {
		fmt.Println(orig, "=>", dest)
	}
}
