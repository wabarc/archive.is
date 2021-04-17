package is

import (
	"testing"
)

func TestWayback(t *testing.T) {
	var got map[string]string

	tests := []struct {
		name string
		urls []string
		got  int
	}{
		{
			name: "Without URLs",
			urls: []string{},
			got:  0,
		},
		{
			name: "Has one invalid URL",
			urls: []string{"foo bar", "https://example.com/"},
			got:  1,
		},
		{
			name: "URLs full matches",
			urls: []string{"https://example.com/", "https://example.org/"},
			got:  2,
		},
	}

	wbrc := &Archiver{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ = wbrc.Wayback(test.urls)
			if len(got) != test.got {
				t.Errorf("got = %d; want %d", len(got), test.got)
			}
			for orig, dest := range got {
				if testing.Verbose() {
					t.Log(orig, "=>", dest)
				}
			}
		})
	}
}

func TestPlayback(t *testing.T) {
	var got map[string]string

	tests := []struct {
		name string
		urls []string
		got  int
	}{
		{
			name: "Without URLs",
			urls: []string{},
			got:  0,
		},
		{
			name: "Has one invalid URL",
			urls: []string{"foo bar", "https://example.com/"},
			got:  1,
		},
		{
			name: "URLs full matches",
			urls: []string{"https://example.com/", "https://example.org/"},
			got:  2,
		},
	}

	wbrc := &Archiver{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ = wbrc.Playback(test.urls)
			if len(got) != test.got {
				t.Errorf("got = %d; want %d", len(got), test.got)
			}
			for orig, dest := range got {
				if testing.Verbose() {
					t.Log(orig, "=>", dest)
				}
			}
		})
	}
}
