package workstation_test

import (
	"testing"

	ws "go.orx.me/apps/butter/internal/workstation"
)

func TestItemIDIncludesProviderOriginAndTenant(t *testing.T) {
	github := ws.ItemID("workspace", ws.GitHubSource(), "42", "ISSUE", "7")
	gitlab, err := ws.GitHubSource().Canonical()
	if err != nil {
		t.Fatal(err)
	}
	gitlab.Provider = "gitlab"
	gitlab.Origin = "https://gitlab.example"
	gitlab, err = gitlab.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	gitlabID := ws.ItemID("workspace", gitlab, "42", "ISSUE", "7")
	tenant := gitlab
	tenant.Provider = "plane"
	tenant.Origin = "https://plane.example"
	tenant.TenantID = "tenant-a"
	tenant, err = tenant.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	tenantID := ws.ItemID("workspace", tenant, "42", "ISSUE", "7")
	if github == gitlabID || github == tenantID || gitlabID == tenantID {
		t.Fatal("provider, origin, and tenant must partition canonical identities")
	}

}
