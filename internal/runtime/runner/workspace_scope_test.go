package runner

import (
	"slices"
	"sort"
	"testing"

	"google.golang.org/adk/agent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// newTestService builds a Service with a populated registry but no ADK
// runtime. Sufficient for exercising the workspace lookup helpers.
func newTestService(agents map[string]string) *Service {
	return NewServiceForTest(agents)
}

func TestHasAgentInWorkspace(t *testing.T) {
	svc := newTestService(map[string]string{
		"a-shared": "ws-a",
		"a-shared-other": "ws-b",
		"b-only":  "ws-b",
	})

	cases := []struct {
		ws    string
		agent string
		want  bool
	}{
		{"ws-a", "a-shared", true},
		{"ws-b", "a-shared", false},      // same name, different workspace → must fail
		{"ws-b", "b-only", true},
		{"ws-a", "b-only", false},
		{"ws-a", "missing", false},
		{"", "b-only", true},             // empty workspace bypasses for system paths
	}

	for _, tc := range cases {
		got := svc.HasAgentInWorkspace(tc.ws, tc.agent)
		if got != tc.want {
			t.Errorf("HasAgentInWorkspace(%q, %q) = %v, want %v", tc.ws, tc.agent, got, tc.want)
		}
	}
}

func TestHasAgentInWorkspace_RejectsDynamicAgentsFromTenant(t *testing.T) {
	// An agent registered without a proto (e.g. system builder) has no
	// workspace binding and must not be reachable from a tenant context.
	svc := &Service{
		agents:      map[string]agent.Agent{"system": nil},
		agentsProto: map[string]*agentsv1.Agent{},
	}
	if svc.HasAgentInWorkspace("ws-a", "system") {
		t.Fatal("workspace caller must not resolve dynamic agents without a workspace binding")
	}
	if !svc.HasAgentInWorkspace("", "system") {
		t.Fatal("admin/system caller (empty ws) should resolve dynamic agent")
	}
}

func TestAgentNamesForWorkspace(t *testing.T) {
	svc := newTestService(map[string]string{
		"a1": "ws-a",
		"a2": "ws-a",
		"b1": "ws-b",
	})

	got := svc.AgentNamesForWorkspace("ws-a")
	sort.Strings(got)
	want := []string{"a1", "a2"}
	if !slices.Equal(got, want) {
		t.Errorf("AgentNamesForWorkspace(ws-a) = %v, want %v", got, want)
	}

	got = svc.AgentNamesForWorkspace("ws-b")
	sort.Strings(got)
	want = []string{"b1"}
	if !slices.Equal(got, want) {
		t.Errorf("AgentNamesForWorkspace(ws-b) = %v, want %v", got, want)
	}

	// Empty workspace returns all (admin/system).
	got = svc.AgentNamesForWorkspace("")
	sort.Strings(got)
	want = []string{"a1", "a2", "b1"}
	if !slices.Equal(got, want) {
		t.Errorf("AgentNamesForWorkspace(\"\") = %v, want %v", got, want)
	}
}
