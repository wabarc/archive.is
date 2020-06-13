package main

import "fmt"

var (
	version = "0.0.1"
	date    = "unknown"
)

func init() {
	fmt.Printf("version: %s\ndate: %s\n\n", version, date)
}
