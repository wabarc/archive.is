# A Golang and Command-Line Interface to Archive.is

This package is a command-line tool named `archive.is` saving webpage to [archive.today](https://archive.today), it also supports imports as a Golang package for a programmatic. Please report all bugs and issues on [Github](https://github.com/wabarc/archive.is/issues).

## Installation

From source (^Go 1.12):

```sh
go get github.com/wabarc/archive.is
```

From [gobinaries.com](https://gobinaries.com):

```sh
curl -sf https://gobinaries.com/wabarc/archive.is/cmd/archive.is | sh
```

From [releases](https://github.com/wabarc/archive.is/releases)

## Usage

### Command-line

```sh
$ archive.is https://www.google.com https://www.bbc.com

Output:
version: 0.0.1
date: unknown

https://www.google.com => https://archive.li/JYVMT
https://www.bbc.com => https://archive.li/HjqQV
```

### Go package interfaces

```go
package main

import (
        "fmt"

        "github.com/wabarc/archive.is/pkg"
)

func main() {
        links := []string{"https://www.google.com", "https://www.bbc.com"}
        arc := &is.Archiver{}
        got, _ := arc.Wayback(links)
        for orig, dest := range got {
                fmt.Println(orig, "=>", dest)
        }
}

// Output:
// https://www.google.com => https://archive.li/JYVMT
// https://www.bbc.com => https://archive.li/HjqQV
```

### Access Tor Hidden Service

[archive.today](https://archive.today) providing [Tor Hidden Service](http://archivecaslytosk.onion/) to saving webpage, and it's preferred to access
Tor Hidden Service, access <http://archive.today> if Tor Hidden Service is unavailable.

By default, the program will dial a proxy using tor socks port `127.0.0.1:9050`,
use `TOR_HOST` and `TOR_SOCKS_PORT` specified a different host and port

It'll look up tor executable file if dial socks proxy failed, and start it to dial proxy.

## FAQ

### archive.today is unavailable?

Archive.today may have enforced a strictly CAPTCHA policy, causing an exception to the request.

Solve:

Find `cf_clearance` item from cookies, and set as system environmental variable `ARCHIVE_COOKIE`,
such as `ARCHIVE_COOKIE=cf_clearance=ab170e4acc49bbnsaff8687212d2cdb987e5b798-1234542375-KDUKCHU`

## License

This software is released under the terms of the GNU General Public License v3.0. See the [LICENSE](https://github.com/wabarc/archive.is/blob/main/LICENSE) file for details.

