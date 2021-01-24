# A Golang and Command-Line Interface to Archive.is

This package is a command-line tool named `archive.is` saving webpage to [archive.today](https://archive.today), it also supports imports as a Golang package for a programmatic. Please report all bugs and issues on [Github](https://github.com/wabarc/archive.is/issues).

## Installation

From source:

```sh
$ go get github.com/wabarc/archive.is
```

From [gobinaries.com](https://gobinaries.com):

```sh
$ curl -sf https://gobinaries.com/wabarc/archive.is | sh
```

From [releases](https://github.com/wabarc/archive.is/releases)

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

### Archive.today is unavailable?

Archive.today may have enforced a strictly CAPTCHA policy, causing an exception to the request.

Solve:

Find `cf_clearance` item from cookies, and set as system environmental variable `ARCHIVE_COOKIE`,
such as `ARCHIVE_COOKIE=cf_clearance=ab170e4acc49bbnsaff8687212d2cdb987e5b798-1234542375-KDUKCHU`

## License

This software is released under the terms of the GNU General Public License v3.0. See the [LICENSE](https://github.com/wabarc/archive.is/blob/main/LICENSE) file for details.

