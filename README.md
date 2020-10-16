# A Golang and Command-Line Interface to Archive.is

This package is a command-line tool named `archive.is` saving webpage to [archive.today](https://archive.today), it also supports imports as a Golang package for a programmatic. Please report all bugs and issues on [Github](https://github.com/wabarc/archive.is/issues).

## Installation

```sh
$ go get github.com/wabarc/archive.is
```

## Usage

#### Command-line

```sh
$ archive.is https://www.google.com https://www.bbc.com

Output:
version: 0.0.1
date: unknown

https://www.google.com => https://archive.li/JYVMT
https://www.bbc.com => https://archive.li/HjqQV
```

#### Go package interfaces

```go
package main

import (
        "fmt"

        "github.com/wabarc/archive.is/pkg"
)

func main() {
        links := []string{"https://www.google.com", "https://www.bbc.com"}
        wbrc := &is.Archiver{}
        got, _ := wbrc.Wayback(links)
        for orig, dest := range got {
                fmt.Println(orig, "=>", dest)
        }
}

// Output:
// https://www.google.com => https://archive.li/JYVMT
// https://www.bbc.com => https://archive.li/HjqQV
```

## FAQ

- Archive.today is unavailable?

Archive.today may have enforced a strictly CAPTCHA policy, causing an exception to the request.

Solve:

Find `cf_clearance` item from cookies, and set as system environmental variable `ARCHIVE_COOKIE`,
such as `ARCHIVE_COOKIE=cf_clearance=ab170e4acc49bbnsaff8687212d2cdb987e5b798-1234542375-KDUKCHU`

## License

Permissive GPL 3.0 license, see the [LICENSE](https://github.com/wabarc/archive.is/blob/master/LICENSE) file for details.

