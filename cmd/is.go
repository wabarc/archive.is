package is

import (
	"flag"
	"fmt"
	"github.com/wabarc/archive.is/pkg"
	"os"
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

	saved := is.Wayback(args)
	for _, link := range saved {
		fmt.Println(link)
	}
}
