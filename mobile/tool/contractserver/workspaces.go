package main

// fixtureWorkspace describes one Workspace the contract fixture creates for the
// signed-in account. Each fixture file registers its own Workspaces from an
// init function, so adding a fixture never edits a shared list.
type fixtureWorkspace struct {
	ID string
	// Owner makes the account an owner (default: member).
	Owner bool
	// NoSeedAgent leaves the Workspace without the default "shared" Agent;
	// its fixture registers whatever it needs.
	NoSeedAgent bool
}

var fixtureWorkspaces = []fixtureWorkspace{{ID: "alpha"}, {ID: "beta"}}

func registerFixtureWorkspaces(ws ...fixtureWorkspace) {
	fixtureWorkspaces = append(fixtureWorkspaces, ws...)
}
