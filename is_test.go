package is

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wabarc/helper"
)

var handler = func(w http.ResponseWriter, r *http.Request) {
	page := "testdata/homepage.html"
	switch {
	case strings.Contains(r.URL.Path, "submit"):
		// wayback or save
	case r.Method == http.MethodGet && r.URL.Path != "/":
		// playback or search
		page = "testdata/search.html"
	}

	buf, err := os.ReadFile(page)
	if err != nil {
		_, _ = w.Write([]byte(``))
	} else {
		_, _ = w.Write(buf)
	}
}

func TestWayback(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handler)

	uri := "https://example.com"
	u, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	arc := NewArchiver(httpClient)
	defer arc.CloseTor()
	_, err = arc.Wayback(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPlayback(t *testing.T) {
	httpClient, mux, server := helper.MockServer()
	defer server.Close()

	mux.HandleFunc("/", handler)

	uri := "https://example.com"
	u, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	arc := NewArchiver(httpClient)
	defer arc.CloseTor()
	_, err = arc.Playback(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
}
