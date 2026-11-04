// Package outbound confines provider requests to an approved HTTPS origin.
package outbound

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewClient checks DNS at every dial and connects to the checked IP itself.
// Transport injection is for tests or an already confined transport decorator.
func NewClient(origin string, transport http.RoundTripper) (*http.Client, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.Contains(u.EscapedPath(), "%") {
		return nil, errors.New("unapproved provider origin")
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	if transport == nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.Proxy = nil
		t.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			host, actualPort, e := net.SplitHostPort(address)
			if e != nil || !strings.EqualFold(host, u.Hostname()) || actualPort != port {
				return nil, errors.New("unapproved destination")
			}
			ips, e := net.DefaultResolver.LookupIPAddr(ctx, host)
			if e != nil {
				return nil, errors.New("DNS unavailable")
			}
			// A mixed public/private answer is rejected, rather than trying the
			// private address after a connection failure or a DNS change.
			for _, ip := range ips {
				if !PublicIP(ip.IP) {
					return nil, errors.New("unapproved network")
				}
			}
			d := net.Dialer{Timeout: 15 * time.Second}
			for _, ip := range ips {
				if conn, e := d.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port)); e == nil {
					return conn, nil
				}
			}
			return nil, errors.New("provider unavailable")
		}
		transport = t
	}
	return &http.Client{Transport: &confined{next: transport, origin: u}, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, nil
}

func PublicIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, cidr := range []string{"0.0.0.0/8", "240.0.0.0/4", "64:ff9b::/96", "64:ff9b:1::/48", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "2001:db8::/32"} {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return false
		}
	}
	return true
}

type confined struct {
	next   http.RoundTripper
	origin *url.URL
}

func (t *confined) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Scheme != t.origin.Scheme || !strings.EqualFold(r.URL.Host, t.origin.Host) || r.URL.User != nil || (r.Method != http.MethodGet && !(r.Method == http.MethodPost && r.URL.Host == "api.linear.app" && r.URL.Path == "/graphql")) {
		return nil, errors.New("unapproved provider request")
	}
	return t.next.RoundTrip(r)
}
