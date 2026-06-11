package main

// serverCommit is set by build ldflags. Keep the default useful for local
// ad-hoc builds where ldflags are omitted.
var serverCommit = "unknown"

func serverBuildCommit() string {
	if serverCommit == "" {
		return "unknown"
	}
	return serverCommit
}
