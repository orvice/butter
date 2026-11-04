package plane

import (
	"context"
	"os"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

// TestRealV2ReadProfile is opt-in and performs only GET operations. It cannot
// replace the human-controlled lifecycle/revocation matrix in the profile doc.
func TestRealV2ReadProfile(t *testing.T) {
	if os.Getenv("BUTTER_TEST_PLANE_ACCEPTANCE") != "1" {
		t.Skip("real Plane acceptance needs an explicitly approved target and credential")
	}
	required := func(name string) string {
		value := os.Getenv(name)
		if value == "" {
			t.Fatalf("required acceptance setting %s is missing", name)
		}
		return value
	}
	api, web := required("BUTTER_TEST_PLANE_API_BASE"), required("BUTTER_TEST_PLANE_WEB_BASE")
	slug, projectID := required("BUTTER_TEST_PLANE_WORKSPACE"), required("BUTTER_TEST_PLANE_PROJECT_ID")
	key := required("BUTTER_TEST_PLANE_API_KEY")
	version := required("BUTTER_TEST_PLANE_VERSION")
	p, err := New(api, web, slug, nil)
	if err != nil {
		t.Fatal("target does not satisfy the approved HTTPS profile")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	identity, err := p.Probe(ctx, key)
	if err != nil {
		t.Fatalf("credential probe category=%s", ws.Classify(err).Category)
	}
	if identity.Source.Provider != "plane" || identity.PrincipalID == "" {
		t.Fatal("invalid probe identity")
	}
	selected, err := p.ResolveScope(ctx, key, projectID)
	if err != nil {
		t.Fatalf("project check category=%s", ws.Classify(err).Category)
	}
	result, err := p.Scan(ctx, key, selected, time.Now().UTC())
	if err != nil {
		t.Fatalf("scan category=%s", ws.Classify(err).Category)
	}
	t.Logf("profile=plane-v2-cursor version=%q read_only=true item_count=%d coverage=%s unresolved=%d", version, len(result.Items), result.Coverage, result.Unresolved)
	if result.Coverage != "COMPLETE" {
		t.Fatal("target qualification remains incomplete: unresolved terminal timestamps")
	}
}
