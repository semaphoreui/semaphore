// Package netproxy dials a TCP address through a SOCKS5 or HTTP proxy. It backs
// the ProxyCommand Semaphore gives to ssh for proxies which are not jump hosts,
// so that no external helper such as netcat has to be installed next to it.
package netproxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

// Credentials authenticate against the proxy itself. They are kept out of the
// proxy URL because that URL is passed on a command line.
type Credentials struct {
	User     string
	Password string
}

// Dial connects to address ("host:port") through the proxy described by
// proxyURL, for example "socks5://proxy.example.org:1080" or
// "http://proxy.example.org:3128".
func Dial(ctx context.Context, proxyURL string, address string, creds Credentials) (net.Conn, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parsing proxy url: %w", err)
	}

	if creds.User != "" {
		u.User = url.UserPassword(creds.User, creds.Password)
	}

	switch u.Scheme {
	case "socks5", "socks5h":
		return dialSOCKS5(ctx, u, address)
	case "http", "https":
		return dialHTTPConnect(ctx, u, address)
	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q", u.Scheme)
	}
}

func dialSOCKS5(ctx context.Context, u *url.URL, address string) (net.Conn, error) {
	var auth *proxy.Auth

	if u.User != nil {
		password, _ := u.User.Password()
		auth = &proxy.Auth{User: u.User.Username(), Password: password}
	}

	dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	// SOCKS5 from x/net implements ContextDialer, so the connect honours the
	// deadline of the caller.
	if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
		return contextDialer.DialContext(ctx, "tcp", address)
	}

	return dialer.Dial("tcp", address)
}

// dialHTTPConnect opens a tunnel with the CONNECT method, which is what an HTTP
// proxy offers for a protocol it does not understand, such as ssh.
func dialHTTPConnect(ctx context.Context, u *url.URL, address string) (net.Conn, error) {
	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		host, _, splitErr := net.SplitHostPort(u.Host)
		if splitErr != nil {
			host = u.Host
		}
		conn = tls.Client(conn, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: address},
		Host:   address,
		Header: http.Header{},
	}

	if u.User != nil {
		password, _ := u.User.Password()
		req.Header.Set("Proxy-Authorization", "Basic "+basicAuth(u.User.Username(), password))
	}

	if err = req.Write(conn); err != nil {
		conn.Close() //nolint: errcheck
		return nil, fmt.Errorf("sending CONNECT: %w", err)
	}

	// The response is read with a bufio.Reader, which may buffer past the
	// headers, so the tunnel keeps reading through it.
	br := bufio.NewReader(conn)

	res, err := http.ReadResponse(br, req)
	if err != nil {
		conn.Close() //nolint: errcheck
		return nil, fmt.Errorf("reading CONNECT response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		conn.Close() //nolint: errcheck
		return nil, fmt.Errorf("proxy refused CONNECT: %s", strings.TrimSpace(res.Status))
	}

	return &bufferedConn{Conn: conn, reader: br}, nil
}

// bufferedConn keeps whatever the CONNECT response reader buffered ahead of the
// tunnel, so no byte of the tunnelled protocol is lost.
type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func basicAuth(user, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
}
