package workstation

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestCanonicalItemIdentityNamespacesAndLegacyStability(t *testing.T) {
	source, err := (SourceIdentity{Provider: "GitHub", Origin: "https://API.GITHUB.COM:443/"}).Canonical()
	if err != nil || source != GitHubSource() {
		t.Fatal(source, err)
	}
	legacy := sha256.Sum256([]byte("ws\x00github\x00https://api.github.com\x0042\x00ISSUE\x001"))
	id := ItemID("ws", source, "42", "ISSUE", "1")
	if id != hex.EncodeToString(legacy[:]) {
		t.Fatal("existing GitHub ID changed")
	}
	variants := []SourceIdentity{{Provider: "gitlab", Origin: source.Origin}, {Provider: source.Provider, Origin: "https://another.example/api/v4"}, {Provider: source.Provider, Origin: source.Origin, TenantID: "org-1"}, {Provider: source.Provider, Origin: source.Origin, TenantID: "org-2"}}
	seen := map[string]bool{id: true}
	for _, variant := range variants {
		next := ItemID("ws", variant, "42", "ISSUE", "1")
		if seen[next] {
			t.Fatal("provider/origin/tenant collision")
		}
		seen[next] = true
	}
	if (Record{}).SourceIdentity() != GitHubSource() {
		t.Fatal("legacy source-less record not supported")
	}
	if Public(Record{Source: variants[0]}).Provider != "gitlab" {
		t.Fatal("public provider hardcoded")
	}
}
