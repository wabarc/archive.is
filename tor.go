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

func newTorClient(ctx context.Context) (client *http.Client, t *tor.Tor, err error) {
	var dialer proxy.ContextDialer
	addr, isUseProxy := useProxy()
	if isUseProxy {
		// Create a socks5 dialer
		pxy, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
		if err != nil {
			return nil, t, fmt.Errorf("Can't connect to the proxy: %w", err)
		}

		dialer = pxy.(interface {
			DialContext(ctx context.Context, network, addr string) (net.Conn, error)
		})
	} else {
		// Lookup tor executable file
		if _, err := exec.LookPath("tor"); err != nil {
			return nil, t, fmt.Errorf("%w", err)
		}

		// Start tor with default config
		startConf := &tor.StartConf{TempDataDirBase: os.TempDir(), RetainTempDataDir: false, NoHush: false}
		t, err = tor.Start(context.TODO(), startConf)
		if err != nil {
			return nil, t, fmt.Errorf("Make connection failed: %w", err)
		}
		// defer t.Close()
		t.DeleteDataDirOnClose = true
		t.StopProcessOnClose = true

		// Wait at most a minute to start network and get
		dialCtx, dialCancel := context.WithTimeout(ctx, time.Minute)
		defer dialCancel()
		// t.ProcessCancelFunc = dialCancel

		// Make connection
		dialer, err = t.Dialer(dialCtx, nil)
		if err != nil {
			t.Close()
			return nil, t, fmt.Errorf("Make connection failed: %w", err)
		}
	}

	return &http.Client{
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
	}, t, nil
}

func closeTor(t *tor.Tor) error {
	if t != nil {
		t.Close()
	}
	return nil
}

func useProxy() (addr string, ok bool) {
	host := os.Getenv("TOR_HOST")
	port := os.Getenv("TOR_SOCKS_PORT")
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "9050"
	}

	addr = net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		logger.Debug("Try to connect tor proxy failed: %v", err)
		return addr, false
	}
	if conn != nil {
		conn.Close()
		logger.Debug("Connected: %v", addr)
		return addr, true
	}

	return addr, false
}
