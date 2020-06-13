package main

import (
	"github.com/wabarc/archive.is/pkg"
	"strings"
	"testing"
)

func TestWayback(t *testing.T) {
	links := []string{"https://www.google.com"}
	got := is.Wayback(links)
	s := strings.Join(got, " ")
	if strings.Contains(s, "archive") == false {
		t.Error(got)
		t.Fail()
	}
}
