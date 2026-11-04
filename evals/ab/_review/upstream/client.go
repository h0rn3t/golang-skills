package upstream

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// NewBackendClient returns the client for unary calls to backends. Backends
// speak HTTP/2 over TLS; the frame size is raised for large responses.
func NewBackendClient(roots *x509.CertPool) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	t := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		TLSClientConfig:       &tls.Config{RootCAs: roots},
		TLSHandshakeTimeout:   5 * time.Second,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		HTTP2:                 &http.HTTP2Config{MaxReadFrameSize: 1 << 20},
	}
	return &http.Client{Transport: t, Timeout: 30 * time.Second}
}

// NewWatchClient returns the client for watch streams: one long-lived
// HTTP/2 connection per backend carries every subscriber's stream, each
// held open for hours.
func NewWatchClient(roots *x509.CertPool) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{RootCAs: roots}
	t.MaxIdleConnsPerHost = 4
	t.ResponseHeaderTimeout = 10 * time.Second
	t.HTTP2 = &http.HTTP2Config{MaxConcurrentStreams: 1000}
	return &http.Client{Transport: t}
}

// Call sends a unary request to the backend routed for prefix and returns
// the response body.
func Call(ctx context.Context, c *http.Client, reg *Registry, prefix, path string) ([]byte, error) {
	b, err := reg.Lookup(prefix)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.URL+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", b.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("call %s: status %d", b.Name, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}
