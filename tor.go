package is

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/wabarc/logger"
	"golang.org/x/net/proxy"
)

func newTorClient(done <-chan bool) (*http.Client, error) {
	var dialer proxy.ContextDialer
	if useProxy() {
		// Create a socks5 dialer
		pxy, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("Can't connect to the proxy: %w", err)
		}

		dialer = pxy.(interface {
			DialContext(ctx context.Context, network, addr string) (net.Conn, error)
		})
	} else {
		// Lookup tor executable file
		if _, err := exec.LookPath("tor"); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		// Start tor with default config
		startConf := &tor.StartConf{TempDataDirBase: os.TempDir()}
		t, err := tor.Start(nil, startConf)
		if err != nil {
			return nil, fmt.Errorf("Make connection failed: %w", err)
		}
		// defer t.Close()

		// Wait at most a minute to start network and get
		dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
		defer dialCancel()

		// Make connection
		dialer, err = t.Dialer(dialCtx, nil)
		if err != nil {
			t.Close()
			return nil, fmt.Errorf("Make connection failed: %w", err)
		}

		go func() {
			// Auto close tor client after 10 min
			tick := time.NewTicker(10 * time.Minute)
			for {
				select {
				case <-done:
					logger.Debug("Closed tor client")
					tick.Stop()
					t.Close()
					return
				case <-tick.C:
					logger.Debug("Closed tor client, timeout")
					tick.Stop()
					t.Close()
					return
				default:
					logger.Debug("Waiting for close tor client")
				}
			}
		}()
	}

	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: noRedirect,
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
	}, nil
}

func useProxy() bool {
	host := os.Getenv("TOR_HOST")
	port := os.Getenv("TOR_SOCKS_PORT")
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "9050"
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Second)
	if err != nil {
		logger.Debug("Try to connect tor proxy failed: %v", err)
		return false
	}
	if conn != nil {
		conn.Close()
		logger.Debug("Connected: %v", net.JoinHostPort(host, port))
		return true
	}

	return false
}
