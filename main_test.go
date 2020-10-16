package main

import (
	"strings"
	"testing"

	"github.com/wabarc/archive.is/pkg"
)

func TestWayback(t *testing.T) {
	links := []string{"https://www.google.com"}
	wbrc := &is.Archiver{}
	got, _ := wbrc.Wayback(links)
	for _, dest := range got {
		if strings.Contains(dest, "Archive.today") == false {
			t.Error(got)
			t.Fail()
		}
	}
}
