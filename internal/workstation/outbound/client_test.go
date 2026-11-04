package outbound

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestNetworkAndOriginConfinement(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "::1", "169.254.169.254", "100.64.0.1", "0.0.0.1", "240.0.0.1", "::ffff:127.0.0.1", "64:ff9b::7f00:1", "2001:db8::1"} {
		if PublicIP(net.ParseIP(raw)) {
			t.Fatal("unapproved network", raw)
		}
	}
	if !PublicIP(net.ParseIP("8.8.8.8")) {
		t.Fatal("public IP rejected")
	}
	calls := 0
	c, err := NewClient("https://approved.example/api", roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"https://other.example/secret"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	response, err := c.Get("https://approved.example/api/items")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if calls != 1 || response.StatusCode != 302 {
		t.Fatal("redirect followed")
	}
	for _, target := range []string{"https://other.example/items", "http://approved.example/items", "https://user@approved.example/items"} {
		if _, err := c.Get(target); err == nil {
			t.Fatal("escaped approved origin", target)
		}
	}
	req, _ := http.NewRequest(http.MethodDelete, "https://approved.example/api/items", nil)
	if _, err := c.Do(req); err == nil {
		t.Fatal("write request escaped")
	}
	if calls != 1 {
		t.Fatal("unapproved request reached transport")
	}
	// The production dialer rejects local DNS/IP targets before opening a socket.
	local, _ := NewClient("https://127.0.0.1", nil)
	if _, err := local.Get("https://127.0.0.1/"); err == nil {
		t.Fatal("private address reached")
	}
}
