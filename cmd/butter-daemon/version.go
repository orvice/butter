package main

// daemonVersion is the semantic version of the daemon client binary. It is
// surfaced at registration so the server-side dashboard can display it.
const daemonVersion = "v0.1.0"

// daemonCommit is set by build ldflags. Keep the default useful for local
// ad-hoc builds where ldflags are omitted.
var daemonCommit = "unknown"

func daemonBuildCommit() string {
	if daemonCommit == "" {
		return "unknown"
	}
	return daemonCommit
}

func daemonBuildVersion() string {
	commit := daemonBuildCommit()
	if commit == "unknown" {
		return daemonVersion
	}
	return daemonVersion + "+" + shortCommit(commit)
}

func shortCommit(commit string) string {
	if len(commit) <= 12 {
		return commit
	}
	return commit[:12]
}
