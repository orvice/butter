package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

// The command reads its database from Butter's config the way the service
// does, and the flags override it.
func TestResolveTarget(t *testing.T) {
	full := writeConfig(t, "butter.yaml", "mongo_uri: \"mongodb://db.internal:27017\"\nmongo_db: \"butter-prod\"\nredis_addr: \"redis:6379\"\n")
	dbOnly := writeConfig(t, "db-only.yaml", "mongo_db: \"butter-staging\"\n")

	for _, c := range []struct {
		name                 string
		env, config, uri, db string
		wantURI, wantDB      string
	}{
		{name: "the service's defaults without a config",
			wantURI: defaultMongoURI, wantDB: defaultMongoDB},
		{name: "the config file", config: full,
			wantURI: "mongodb://db.internal:27017", wantDB: "butter-prod"},
		{name: "the service's config file variable", env: full,
			wantURI: "mongodb://db.internal:27017", wantDB: "butter-prod"},
		{name: "--config wins over the variable", env: full, config: dbOnly,
			wantURI: defaultMongoURI, wantDB: "butter-staging"},
		{name: "flags override the config", config: full, uri: "mongodb://snapshot:27017", db: "restored",
			wantURI: "mongodb://snapshot:27017", wantDB: "restored"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(envConfigPath, c.env)
			got, err := resolveTarget(c.config, c.uri, c.db)
			if err != nil {
				t.Fatalf("resolveTarget: %v", err)
			}
			if got.uri != c.wantURI || got.db != c.wantDB {
				t.Errorf("target = %s / %s, want %s / %s", got.uri, got.db, c.wantURI, c.wantDB)
			}
		})
	}
}

// A config file that cannot be read stops the command rather than falling
// back to the defaults.
func TestResolveTargetRefusesAConfigItCannotRead(t *testing.T) {
	t.Setenv(envConfigPath, "")
	if _, err := resolveTarget(filepath.Join(t.TempDir(), "missing.yaml"), "", ""); err == nil {
		t.Error("a missing config file was accepted")
	}
	if _, err := resolveTarget(writeConfig(t, "bad.yaml", "mongo_uri: [unclosed\n"), "", ""); err == nil {
		t.Error("an unparsable config file was accepted")
	}
}

func TestHelpCarriesTheProductionWarning(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := cli([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("--help exit code %d", code)
	}
	help := strings.Join(strings.Fields(stderr.String()), " ")
	for _, want := range []string{
		"WARNING",
		"Never run it against production without the project owner's explicit confirmation, given at the time, after they reviewed a dry run, and after a backup or snapshot of the database.",
		"Without --confirm it is a dry run",
		"-confirm",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("help lacks %q:\n%s", want, stderr.String())
		}
	}
}

func TestBadUsageExitsWithTwo(t *testing.T) {
	for _, args := range [][]string{{"--yes"}, {"web-chat"}} {
		var stdout, stderr bytes.Buffer
		if code := cli(args, &stdout, &stderr); code != 2 {
			t.Errorf("%v: exit code %d, want 2", args, code)
		}
		if stdout.Len() != 0 {
			t.Errorf("%v: wrote to stdout:\n%s", args, stdout.String())
		}
	}
}
