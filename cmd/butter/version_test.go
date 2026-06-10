package main

import "testing"

func TestServerBuildCommit(t *testing.T) {
	oldCommit := serverCommit
	defer func() {
		serverCommit = oldCommit
	}()

	serverCommit = "1234567890abcdef"
	if got, want := serverBuildCommit(), "1234567890abcdef"; got != want {
		t.Fatalf("serverBuildCommit() = %q, want %q", got, want)
	}

	serverCommit = ""
	if got, want := serverBuildCommit(), "unknown"; got != want {
		t.Fatalf("serverBuildCommit() = %q, want %q", got, want)
	}
}
