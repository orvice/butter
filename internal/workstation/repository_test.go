package workstation_test

import (
	"testing"

	ws "go.orx.me/apps/butter/internal/workstation"
)

func TestItemIDIncludesProviderOriginAndTenant(t *testing.T) {
	github := ws.ItemIDFor("workspace", "github", "https://api.github.com", "", "42", "ISSUE", "7")
	gitlab := ws.ItemIDFor("workspace", "gitlab", "https://gitlab.example", "", "42", "ISSUE", "7")
	tenant := ws.ItemIDFor("workspace", "plane", "https://plane.example", "tenant-a", "42", "ISSUE", "7")
	if github == gitlab || github == tenant || gitlab == tenant {
		t.Fatal("provider, origin, and tenant must partition canonical identities")
	}
	if github != ws.ItemID("workspace", "42", "ISSUE", "7") {
		t.Fatal("legacy GitHub identity changed")
	}
}
