package is

import (
	"testing"
)

func TestWayback(t *testing.T) {
	var (
		links []string
		got   map[string]string
	)

	wbrc := &Archiver{}
	got, _ = wbrc.Wayback(links)
	if len(got) != 0 {
		t.Errorf("got = %d; want 0", len(got))
	}

	links = []string{"https://www.bbc.com/", "https://www.google.com/"}
	got, _ = wbrc.Wayback(links)
	if len(got) == 0 {
		t.Errorf("got = %d; want greater than 0", len(got))
	}

	for orig, dest := range got {
		t.Log(orig, "=>", dest)
	}
}
