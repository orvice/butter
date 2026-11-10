package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// countingCloser stands in for a toolset that owns a live connection.
type countingCloser struct {
	closed int
	err    error
}

func (c *countingCloser) Close() error {
	c.closed++
	return c.err
}

func TestResourcesCloseClosesEveryAddedCloser(t *testing.T) {
	first, second := &countingCloser{}, &countingCloser{}
	res := &Resources{}
	res.Add(first, second)

	if err := res.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if first.closed != 1 || second.closed != 1 {
		t.Fatalf("closed = %d, %d; want 1, 1", first.closed, second.closed)
	}

	// Closing again must not close the connections a second time.
	if err := res.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if first.closed != 1 || second.closed != 1 {
		t.Fatalf("closed after second Close = %d, %d; want 1, 1", first.closed, second.closed)
	}
}

func TestResourcesCloseWaitsForTheLastRelease(t *testing.T) {
	closer := &countingCloser{}
	res := &Resources{}
	res.Add(closer)

	release, ok := res.Acquire()
	if !ok {
		t.Fatal("Acquire on fresh resources failed")
	}

	// Retiring while a turn runs must not pull the connection away from it.
	if err := res.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if closer.closed != 0 {
		t.Fatalf("closed %d times while a run still held the resources", closer.closed)
	}
	if _, ok := res.Acquire(); ok {
		t.Fatal("retired resources accepted a new run")
	}

	release()
	if closer.closed != 1 {
		t.Fatalf("closed = %d after the last release, want 1", closer.closed)
	}
}

func TestResourcesAddAfterCloseClosesAtOnce(t *testing.T) {
	closer := &countingCloser{}
	res := &Resources{}
	if err := res.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	res.Add(closer)
	if closer.closed != 1 {
		t.Fatalf("closed = %d, want 1: a late registration must not leak", closer.closed)
	}
}

func TestResourcesReportsCloseErrors(t *testing.T) {
	boom := errors.New("boom")
	res := &Resources{}
	res.Add(&countingCloser{err: boom}, &countingCloser{})
	// A nil Resources is a valid, empty owner.
	var empty *Resources
	empty.Add(&countingCloser{})

	if err := res.Close(); !errors.Is(err, boom) {
		t.Fatalf("Close error = %v, want %v", err, boom)
	}
	if got := res.Err(); !errors.Is(got, boom) {
		t.Fatalf("Err = %v, want %v", got, boom)
	}
	if err := empty.Close(); err != nil {
		t.Fatalf("nil Resources Close: %v", err)
	}
	if release, ok := empty.Acquire(); !ok {
		t.Fatal("nil Resources refused a run")
	} else {
		release()
	}
}

// closeableToolset is a built-in toolset whose owner must close it.
type closeableToolset struct {
	name   string
	closed int
}

func (t *closeableToolset) Name() string { return t.name }

func (t *closeableToolset) Tools(agent.ReadonlyContext) ([]tool.Tool, error) { return nil, nil }

func (t *closeableToolset) Close() error {
	t.closed++
	return nil
}

// TestNewFromProtoClosesToolsetsOfAFailedTree: children are built before their
// parent, so a child that fails must not leave the toolsets of the children
// built before it open.
func TestNewFromProtoClosesToolsetsOfAFailedTree(t *testing.T) {
	opened := map[string]*closeableToolset{}
	factory := func(_ context.Context, pb *agentsv1.Agent) ([]tool.Toolset, error) {
		ts := &closeableToolset{name: pb.GetName() + "-tools"}
		opened[pb.GetName()] = ts
		return []tool.Toolset{ts}, nil
	}

	good := &agentsv1.Agent{AgentId: "good", Name: "good", Config: &agentsv1.AgentConfig{Model: "m1"}}
	// A CURSOR child without a box builder fails after good was built.
	bad := &agentsv1.Agent{AgentId: "bad", Name: "bad", Type: agentsv1.AgentType_AGENT_TYPE_CURSOR}
	root := &agentsv1.Agent{
		AgentId:       "root",
		Name:          "root",
		Config:        &agentsv1.AgentConfig{Model: "m1"},
		ChildAgentIds: []string{"good", "bad"},
	}

	_, _, err := NewFromProtoWithToolsetFactory(t.Context(), root, workflowProviders(), nil, nil, nil, nil, factory, nil,
		AgentPool{"good": good, "bad": bad})
	if err == nil {
		t.Fatal("NewFromProto accepted a CURSOR child without a box builder")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Fatalf("error = %v, want the failing child named", err)
	}
	built, ok := opened["good"]
	if !ok {
		t.Fatal("the first child's toolsets were never built")
	}
	if built.closed != 1 {
		t.Fatalf("first child's toolset closed %d times after the failed build, want 1", built.closed)
	}
}

// TestNewFromProtoOwnsToolsetConnections: a successful build hands back the
// owner of every closeable toolset in the tree.
func TestNewFromProtoOwnsToolsetConnections(t *testing.T) {
	var built *closeableToolset
	factory := func(_ context.Context, _ *agentsv1.Agent) ([]tool.Toolset, error) {
		built = &closeableToolset{name: "tools"}
		return []tool.Toolset{built}, nil
	}
	root := &agentsv1.Agent{AgentId: "root", Name: "root", Config: &agentsv1.AgentConfig{Model: "m1"}}

	_, res, err := NewFromProtoWithToolsetFactory(t.Context(), root, workflowProviders(), nil, nil, nil, nil, factory, nil)
	if err != nil {
		t.Fatalf("NewFromProto: %v", err)
	}
	if res == nil {
		t.Fatal("NewFromProto returned no resources")
	}
	if built == nil {
		t.Fatal("the toolset factory never ran")
	}
	if built.closed != 0 {
		t.Fatalf("toolset closed %d times before its owner", built.closed)
	}
	if err := res.Close(); err != nil {
		t.Fatalf("closing resources: %v", err)
	}
	if built.closed != 1 {
		t.Fatalf("toolset closed %d times, want 1", built.closed)
	}
}
