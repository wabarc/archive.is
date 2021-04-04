package is

import (
	"context"
	"crypto/tls"
	"fmt"
	// "net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/cretz/bine/tor"
	// "golang.org/x/net/proxy"
)

func newTorClient() (*http.Client, *tor.Tor, error) {
	// Lookup tor executable file
	if _, err := exec.LookPath("tor"); err != nil {
		return nil, nil, fmt.Errorf("%w", err)
	}

	// Start tor with default config
	startConf := &tor.StartConf{TempDataDirBase: os.TempDir()}
	t, err := tor.Start(nil, startConf)
	if err != nil {
		return nil, nil, fmt.Errorf("Make connection failed: %w", err)
	}
	// defer t.Close()

	// Wait at most a minute to start network and get
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dialCancel()

	// Make connection
	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		t.Close()
		return nil, nil, fmt.Errorf("Make connection failed: %w", err)
	}

	// Create a socks5 dialer
	// pxy, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	// if err != nil {
	// 	return nil, fmt.Errorf("Can't connect to the proxy: %w", err)
	// }

	// dialer := pxy.(interface {
	// 	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
	// })

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}, t, nil
}
