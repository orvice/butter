package prefs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMissingFileIsZero(t *testing.T) {
	d, err := Load(filepath.Join(t.TempDir(), "nope", "desktop.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if d != (Data{}) {
		t.Fatalf("data = %+v, want zero", d)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "butter", "desktop.json")
	want := Data{Server: "https://butter.example.com", Username: "lion", WorkspaceID: "w1"}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("permissions = %o, want 600", perm)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"password", "token", "secret"} {
		if strings.Contains(string(b), forbidden) {
			t.Fatalf("prefs file carries %q: %s", forbidden, b)
		}
	}
}

func TestSaveReplacesAtomicallyAndLeavesNoTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "desktop.json")
	if err := Save(path, Data{Server: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, Data{Server: "b"}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("left files behind: %v", entries)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Server != "b" {
		t.Fatalf("server = %q, want b", got.Server)
	}
}
