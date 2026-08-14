// Package repotest is a conformance suite for configrepo.AgentRepository
// implementations. (workspace_id, agent_id) is the logical primary key of the
// Agent repository (issue #241); the memory and mongo backends must behave
// identically through this seam.
package repotest

import (
	"context"
	"errors"
	"testing"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory returns a fresh, empty repository for one subtest.
type Factory func(t *testing.T) configrepo.AgentRepository

func newAgent(id, name string) *agentsv1.Agent {
	return &agentsv1.Agent{
		AgentId:     id,
		Name:        name,
		DisplayName: "Agent " + id,
		Description: "test agent " + id,
	}
}

func create(t *testing.T, repo configrepo.AgentRepository, ws string, a *agentsv1.Agent) *agentsv1.Agent {
	t.Helper()
	created, err := repo.CreateAgent(context.Background(), ws, a)
	if err != nil {
		t.Fatalf("CreateAgent %s/%s: %v", ws, a.GetAgentId(), err)
	}
	return created
}

// RunAgents exercises the ID-keyed AgentRepository contract against one
// implementation.
func RunAgents(t *testing.T, factory Factory) {
	ctx := context.Background()

	t.Run("CRUD_by_agent_id", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws", newAgent("writer", "Writer"))

		got, err := repo.GetAgent(ctx, "ws", "writer")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if got.GetName() != "Writer" || got.GetWorkspaceId() != "ws" {
			t.Fatalf("GetAgent = %v, want name Writer in ws", got)
		}

		got.Description = "updated"
		if _, err := repo.UpdateAgent(ctx, "ws", got); err != nil {
			t.Fatalf("UpdateAgent: %v", err)
		}
		got, err = repo.GetAgent(ctx, "ws", "writer")
		if err != nil {
			t.Fatalf("GetAgent after update: %v", err)
		}
		if got.GetDescription() != "updated" {
			t.Fatalf("description = %q, want updated", got.GetDescription())
		}

		if err := repo.DeleteAgent(ctx, "ws", "writer"); err != nil {
			t.Fatalf("DeleteAgent: %v", err)
		}
		if _, err := repo.GetAgent(ctx, "ws", "writer"); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("GetAgent after delete: %v, want ErrNotFound", err)
		}
	})

	t.Run("name_never_selects_a_record", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws", newAgent("writer", "Writer"))
		if _, err := repo.GetAgent(ctx, "ws", "Writer"); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("GetAgent by runtime name: %v, want ErrNotFound", err)
		}
		if err := repo.DeleteAgent(ctx, "ws", "Writer"); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("DeleteAgent by runtime name: %v, want ErrNotFound", err)
		}
	})

	t.Run("empty_agent_id_rejected", func(t *testing.T) {
		repo := factory(t)
		if _, err := repo.CreateAgent(ctx, "ws", newAgent("", "NoID")); !errors.Is(err, configrepo.ErrMissingAgentID) {
			t.Fatalf("CreateAgent without id: %v, want ErrMissingAgentID", err)
		}
		if _, err := repo.GetAgent(ctx, "ws", ""); !errors.Is(err, configrepo.ErrMissingAgentID) {
			t.Fatalf("GetAgent with empty id: %v, want ErrMissingAgentID", err)
		}
		if _, err := repo.UpdateAgent(ctx, "ws", newAgent("", "NoID")); !errors.Is(err, configrepo.ErrMissingAgentID) {
			t.Fatalf("UpdateAgent without id: %v, want ErrMissingAgentID", err)
		}
		if _, err := repo.UpdateAgentCAS(ctx, "ws", newAgent("", "NoID"), 0); !errors.Is(err, configrepo.ErrMissingAgentID) {
			t.Fatalf("UpdateAgentCAS without id: %v, want ErrMissingAgentID", err)
		}
		if err := repo.DeleteAgent(ctx, "ws", ""); !errors.Is(err, configrepo.ErrMissingAgentID) {
			t.Fatalf("DeleteAgent with empty id: %v, want ErrMissingAgentID", err)
		}
	})

	t.Run("duplicate_agent_id_rejected", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws", newAgent("writer", "Writer"))
		if _, err := repo.CreateAgent(ctx, "ws", newAgent("writer", "Other")); !errors.Is(err, configrepo.ErrAlreadyExists) {
			t.Fatalf("CreateAgent duplicate id: %v, want ErrAlreadyExists", err)
		}
	})

	t.Run("duplicate_runtime_name_rejected", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws", newAgent("writer", "Writer"))
		if _, err := repo.CreateAgent(ctx, "ws", newAgent("other", "Writer")); !errors.Is(err, configrepo.ErrAlreadyExists) {
			t.Fatalf("CreateAgent duplicate name: %v, want ErrAlreadyExists", err)
		}

		create(t, repo, "ws", newAgent("editor", "Editor"))
		clash := newAgent("editor", "Writer")
		if _, err := repo.UpdateAgent(ctx, "ws", clash); !errors.Is(err, configrepo.ErrAlreadyExists) {
			t.Fatalf("UpdateAgent onto taken name: %v, want ErrAlreadyExists", err)
		}
	})

	t.Run("update_missing_returns_not_found", func(t *testing.T) {
		repo := factory(t)
		if _, err := repo.UpdateAgent(ctx, "ws", newAgent("ghost", "Ghost")); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("UpdateAgent missing: %v, want ErrNotFound", err)
		}
	})

	t.Run("update_preserves_identity_when_name_changes", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws", newAgent("writer", "Writer"))
		renamed := newAgent("writer", "Renamed")
		if _, err := repo.UpdateAgent(ctx, "ws", renamed); err != nil {
			t.Fatalf("UpdateAgent rename: %v", err)
		}
		got, err := repo.GetAgent(ctx, "ws", "writer")
		if err != nil {
			t.Fatalf("GetAgent after rename: %v", err)
		}
		if got.GetName() != "Renamed" {
			t.Fatalf("name = %q, want Renamed", got.GetName())
		}
	})

	t.Run("cas_bumps_version_and_detects_conflicts", func(t *testing.T) {
		repo := factory(t)
		a := newAgent("writer", "Writer")
		a.Version = 1
		create(t, repo, "ws", a)

		if _, err := repo.UpdateAgentCAS(ctx, "ws", a, 7); !errors.Is(err, configrepo.ErrVersionConflict) {
			t.Fatalf("UpdateAgentCAS wrong version: %v, want ErrVersionConflict", err)
		}

		updated, err := repo.UpdateAgentCAS(ctx, "ws", a, 1)
		if err != nil {
			t.Fatalf("UpdateAgentCAS: %v", err)
		}
		if updated.GetVersion() != 2 {
			t.Fatalf("version = %d, want 2", updated.GetVersion())
		}

		if _, err := repo.UpdateAgentCAS(ctx, "ws", a, 1); !errors.Is(err, configrepo.ErrVersionConflict) {
			t.Fatalf("UpdateAgentCAS stale version: %v, want ErrVersionConflict", err)
		}
		if _, err := repo.UpdateAgentCAS(ctx, "ws", newAgent("ghost", "Ghost"), 0); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("UpdateAgentCAS missing: %v, want ErrNotFound", err)
		}
	})

	t.Run("workspace_isolation", func(t *testing.T) {
		repo := factory(t)
		create(t, repo, "ws-a", newAgent("writer", "WriterA"))
		create(t, repo, "ws-b", newAgent("writer", "WriterB"))

		a, err := repo.GetAgent(ctx, "ws-a", "writer")
		if err != nil {
			t.Fatalf("GetAgent ws-a: %v", err)
		}
		if a.GetName() != "WriterA" {
			t.Fatalf("ws-a agent name = %q, want WriterA", a.GetName())
		}

		if _, err := repo.GetAgent(ctx, "ws-c", "writer"); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("GetAgent other workspace: %v, want ErrNotFound", err)
		}

		if err := repo.DeleteAgent(ctx, "ws-a", "writer"); err != nil {
			t.Fatalf("DeleteAgent ws-a: %v", err)
		}
		if _, err := repo.GetAgent(ctx, "ws-b", "writer"); err != nil {
			t.Fatalf("ws-b agent must survive ws-a delete: %v", err)
		}

		all, err := repo.ListAgentsAcrossWorkspaces(ctx)
		if err != nil {
			t.Fatalf("ListAgentsAcrossWorkspaces: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("across workspaces = %d agents, want 1", len(all))
		}

		inB, err := repo.ListAgents(ctx, "ws-b")
		if err != nil {
			t.Fatalf("ListAgents ws-b: %v", err)
		}
		if len(inB) != 1 {
			t.Fatalf("ws-b = %d agents, want 1", len(inB))
		}
	})

	t.Run("tombstones_stay_readable_and_id_stays_reserved", func(t *testing.T) {
		repo := factory(t)
		a := newAgent("writer", "Writer")
		a.Version = 1
		create(t, repo, "ws", a)

		tomb := newAgent("writer", "Writer")
		tomb.LifecycleStatus = agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED
		if _, err := repo.UpdateAgentCAS(ctx, "ws", tomb, 1); err != nil {
			t.Fatalf("tombstone via CAS: %v", err)
		}

		got, err := repo.GetAgent(ctx, "ws", "writer")
		if err != nil {
			t.Fatalf("GetAgent tombstone: %v", err)
		}
		if got.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
			t.Fatalf("lifecycle = %v, want DELETED", got.GetLifecycleStatus())
		}

		// The tombstone still holds the ID: creating a new agent with it fails.
		if _, err := repo.CreateAgent(ctx, "ws", newAgent("writer", "Reborn")); !errors.Is(err, configrepo.ErrAlreadyExists) {
			t.Fatalf("CreateAgent over tombstone: %v, want ErrAlreadyExists", err)
		}
	})
}
