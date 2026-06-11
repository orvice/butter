package main

import "testing"

func TestDaemonBuildVersionIncludesShortCommit(t *testing.T) {
	oldCommit := daemonCommit
	defer func() {
		daemonCommit = oldCommit
	}()

	daemonCommit = "1234567890abcdef"
	if got, want := daemonBuildVersion(), "v0.1.0+1234567890ab"; got != want {
		t.Fatalf("daemonBuildVersion() = %q, want %q", got, want)
	}
	if got, want := daemonBuildCommit(), "1234567890abcdef"; got != want {
		t.Fatalf("daemonBuildCommit() = %q, want %q", got, want)
	}
}

func TestDaemonBuildVersionOmitsUnknownCommit(t *testing.T) {
	oldCommit := daemonCommit
	defer func() {
		daemonCommit = oldCommit
	}()

	daemonCommit = ""
	if got, want := daemonBuildVersion(), "v0.1.0"; got != want {
		t.Fatalf("daemonBuildVersion() = %q, want %q", got, want)
	}
	if got, want := daemonBuildCommit(), "unknown"; got != want {
		t.Fatalf("daemonBuildCommit() = %q, want %q", got, want)
	}
}
