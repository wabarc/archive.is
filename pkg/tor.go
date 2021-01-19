package is

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/cretz/bine/tor"
)

func (arc *Archiver) dialTor() (*tor.Tor, error) {
	// Lookup tor executable file
	if _, err := exec.LookPath("tor"); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// Start tor with default config
	t, err := tor.Start(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Make connection failed: %w", err)
	}
	// defer t.Close()

	// Wait at most a minute to start network and get
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dialCancel()

	// Make connection
	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		t.Close()
		return nil, fmt.Errorf("Make connection failed: %w", err)
	}

	arc.DialContext = dialer.DialContext
	arc.SkipTLSVerification = true

	return t, nil
}
