# A Golang and Command-Line Interface to Archive.is

This package installs a command-line tool named archive.is for using Archive.is from the command-line. It also installs the Golang package for programmatic snapshot webpage to archive.is. Please report all bugs and issues on [Github](https://github.com/wabarc/archive.is/issues).

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

3.21s  119488 https://www.google.com
2.75s  836849 https://www.bbc.com
5.96s elapsed

https://archive.li/JYVMT
https://archive.li/HjqQV
```

#### Go package interfaces

```go
package main

import (
	"fmt"
	"github.com/wabarc/archive.is/pkg"
	"strings"
)

func main() {
	links := []string{"https://www.google.com", "https://www.bbc.com"}
	r := ia.Wayback(links)
	fmt.Println(strings.Join(r, "\n"))
}

// Output:
// 2.45s  119574 https://www.google.com
// 1.31s  836935 https://www.bbc.com
// 3.76s elapsed
//
// https://archive.li/JYVMT
// https://archive.li/HjqQV
```

## License

Permissive GPL 3.0 license, see the [LICENSE](https://github.com/wabarc/archive.is/blob/master/LICENSE) file for details.

