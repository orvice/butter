package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup"
	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup/webchatcleanuptest"
)

// The command itself: a dry run by default, and --confirm to delete. The
// cleanup's own tests are in internal/maintenance/webchatcleanup.
func TestCommandIsADryRunUnlessConfirmed(t *testing.T) {
	db := webchatcleanuptest.DB(t)
	webchatcleanuptest.Seed(t, db)
	uri := os.Getenv(webchatcleanuptest.URIEnv)
	t.Setenv(envConfigPath, "")
	webChatSessions := func() int64 {
		t.Helper()
		n, err := db.Collection(webchatcleanup.SessionsCollection).CountDocuments(t.Context(), bson.M{"app_name": webchatcleanup.App})
		if err != nil {
			t.Fatalf("count sessions: %v", err)
		}
		return n
	}

	var stdout, stderr bytes.Buffer
	if code := cli([]string{"--mongo-uri", uri, "--mongo-db", db.Name()}, &stdout, &stderr); code != 0 {
		t.Fatalf("dry run exit code %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), fmt.Sprintf("DRY RUN on database %q", db.Name())) {
		t.Errorf("dry run does not name its database:\n%s", stdout.String())
	}
	if n := webChatSessions(); n != webchatcleanuptest.WantTotal.Sessions {
		t.Fatalf("after a dry run %d web-chat sessions remain, want %d", n, webchatcleanuptest.WantTotal.Sessions)
	}

	stdout.Reset()
	if code := cli([]string{"--mongo-uri", uri, "--mongo-db", db.Name(), "--confirm"}, &stdout, &stderr); code != 0 {
		t.Fatalf("confirmed run exit code %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "4. adk_sessions: deleted 6") {
		t.Errorf("confirmed run output:\n%s", stdout.String())
	}
	if n := webChatSessions(); n != 0 {
		t.Fatalf("after --confirm %d web-chat sessions remain", n)
	}
}
