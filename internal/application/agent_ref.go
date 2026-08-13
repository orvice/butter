package application

import (
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// equalSubAgents reports whether two embedded sub-agent trees are identical —
// used by UpdateAgent to allow legacy agents to round-trip unchanged while
// rejecting embedded-tree mutations (the write path was removed in V2).
func equalSubAgents(a, b []*agentsv1.Agent) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !proto.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
