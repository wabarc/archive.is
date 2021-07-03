package is

import (
	"context"
	"net/url"
	"testing"
)

func TestWayback(t *testing.T) {
	uri := "https://example.com"
	u, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	wbrc := &Archiver{}
	_, err = wbrc.Wayback(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPlayback(t *testing.T) {
	uri := "https://example.com"
	u, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	wbrc := &Archiver{}
	_, err = wbrc.Playback(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
}
