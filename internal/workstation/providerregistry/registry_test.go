package providerregistry

import (
	ws "go.orx.me/apps/butter/internal/workstation"
	"net/http"
	"testing"
)

type transport struct{}

func (transport) RoundTrip(*http.Request) (*http.Response, error) { return nil, nil }
func TestAllKindsSelectProviderContract(t *testing.T) {
	for _, kind := range []Kind{GitHub, GitLab, Linear, Plane} {
		p, e := New(Config{Kind: kind}, transport{})
		if e != nil {
			t.Fatalf("%s: %v", kind, e)
		}
		if _, ok := p.(ws.Provider); !ok {
			t.Fatalf("%s does not implement provider contract", kind)
		}
	}
}
func TestHostAndProfileFailClosed(t *testing.T) {
	cases := []Config{{Kind: GitHub, Host: "https://evil.example"}, {Kind: GitLab, Profile: "self_hosted"}, {Kind: Linear, Host: "https://linear.example"}, {Kind: Plane, Profile: "self_hosted"}, {Kind: Plane, Profile: "self_hosted", Host: "http://plane.local"}, {Kind: "unknown"}}
	for _, c := range cases {
		if _, e := New(c, transport{}); e == nil {
			t.Fatalf("accepted unsafe config: %+v", c)
		}
	}
	if _, e := New(Config{Kind: Plane, Profile: "self_hosted", Host: "https://plane.example"}, transport{}); e != nil {
		t.Fatal(e)
	}
}
func TestCloudAliasesRemainExplicit(t *testing.T) {
	if _, e := New(Config{Kind: Plane, Profile: "cloud", Host: "https://api.plane.so"}, transport{}); e != nil {
		t.Fatal(e)
	}
	if _, e := New(Config{Kind: Plane, Profile: "cloud", Host: "https://api.plane.so/"}, transport{}); e != nil {
		t.Fatal(e)
	}
}
