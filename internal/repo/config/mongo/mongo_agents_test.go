package mongo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/repotest"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// testDB connects to the MongoDB named by BUTTER_TEST_MONGO_URI and hands the
// test a throwaway database. Without the env var the test is skipped, so the
// default test run needs no infrastructure.
func testDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI not set; skipping mongo integration test")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	db := client.Database(fmt.Sprintf("butter_config_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func newStore(t *testing.T, db *mongo.Database) *Store {
	t.Helper()
	store := New(db)
	if err := store.EnsureIndexes(context.Background()); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	return store
}

func TestMongoAgentRepositoryConformance(t *testing.T) {
	repotest.RunAgents(t, func(t *testing.T) configrepo.AgentRepository {
		return newStore(t, testDB(t))
	})
}

func TestMongoModelProviderRepositoryConformance(t *testing.T) {
	repotest.RunModelProviders(t, func(t *testing.T) configrepo.ModelProviderRepository {
		return newStore(t, testDB(t))
	})
}

// seedLegacyAgentDoc inserts an agent document exactly as the pre-#241 code
// wrote it: _id = "workspace_id:name".
func seedLegacyAgentDoc(t *testing.T, s *Store, ws string, agent *agentsv1.Agent) string {
	t.Helper()
	spec, err := marshal(agent)
	if err != nil {
		t.Fatalf("marshal legacy agent: %v", err)
	}
	legacyID := compositeID(ws, agent.GetName())
	doc := configDoc{ID: legacyID, WorkspaceID: ws, Name: agent.GetName(), Spec: spec, AgentID: agent.GetAgentId(), Version: agent.GetVersion()}
	if _, err := s.agents.InsertOne(context.Background(), doc); err != nil {
		t.Fatalf("seed legacy doc: %v", err)
	}
	return legacyID
}

func physicalID(t *testing.T, s *Store, ws, agentID string) string {
	t.Helper()
	var doc configDoc
	if err := s.agents.FindOne(context.Background(), agentFilter(ws, agentID)).Decode(&doc); err != nil {
		t.Fatalf("find agent doc: %v", err)
	}
	return doc.ID
}

// TestLegacyIDDocumentsSurviveWrites proves a document whose _id is the
// historical "workspace_id:name" composite stays fully readable and writable
// through the ID-keyed CRUD without its _id ever being rewritten.
func TestLegacyIDDocumentsSurviveWrites(t *testing.T) {
	s := newStore(t, testDB(t))
	ctx := context.Background()

	agent := &agentsv1.Agent{AgentId: "writer", Name: "Writer", WorkspaceId: "ws", Description: "legacy", Version: 1}
	legacyID := seedLegacyAgentDoc(t, s, "ws", agent)

	got, err := s.GetAgent(ctx, "ws", "writer")
	if err != nil {
		t.Fatalf("GetAgent legacy doc: %v", err)
	}
	if got.GetDescription() != "legacy" {
		t.Fatalf("description = %q, want legacy", got.GetDescription())
	}

	got.Description = "updated"
	if _, err := s.UpdateAgent(ctx, "ws", got); err != nil {
		t.Fatalf("UpdateAgent legacy doc: %v", err)
	}
	if id := physicalID(t, s, "ws", "writer"); id != legacyID {
		t.Fatalf("_id after UpdateAgent = %q, want preserved %q", id, legacyID)
	}

	got.Description = "cas"
	if _, err := s.UpdateAgentCAS(ctx, "ws", got, 1); err != nil {
		t.Fatalf("UpdateAgentCAS legacy doc: %v", err)
	}
	if id := physicalID(t, s, "ws", "writer"); id != legacyID {
		t.Fatalf("_id after UpdateAgentCAS = %q, want preserved %q", id, legacyID)
	}

	// Renaming through the ID-keyed update also keeps the (now name-shaped
	// but opaque) physical identifier.
	got, err = s.GetAgent(ctx, "ws", "writer")
	if err != nil {
		t.Fatalf("GetAgent after CAS: %v", err)
	}
	got.Name = "RenamedWriter"
	if _, err := s.UpdateAgent(ctx, "ws", got); err != nil {
		t.Fatalf("UpdateAgent rename legacy doc: %v", err)
	}
	if id := physicalID(t, s, "ws", "writer"); id != legacyID {
		t.Fatalf("_id after rename = %q, want preserved %q", id, legacyID)
	}

	if err := s.DeleteAgent(ctx, "ws", "writer"); err != nil {
		t.Fatalf("DeleteAgent legacy doc: %v", err)
	}
}

// TestLegacyCASTreatsMissingVersionAsZero covers rows written before the
// promoted version field existed: they carry no "version" key at all.
func TestLegacyCASTreatsMissingVersionAsZero(t *testing.T) {
	s := newStore(t, testDB(t))
	ctx := context.Background()

	agent := &agentsv1.Agent{AgentId: "writer", Name: "Writer", WorkspaceId: "ws"}
	legacyID := seedLegacyAgentDoc(t, s, "ws", agent)
	if _, err := s.agents.UpdateOne(ctx, bson.M{"_id": legacyID}, bson.M{"$unset": bson.M{"version": ""}}); err != nil {
		t.Fatalf("unset version: %v", err)
	}

	updated, err := s.UpdateAgentCAS(ctx, "ws", agent, 0)
	if err != nil {
		t.Fatalf("UpdateAgentCAS on version-less row: %v", err)
	}
	if updated.GetVersion() != 1 {
		t.Fatalf("version = %d, want 1", updated.GetVersion())
	}
}

// TestNewDocumentIDsAvoidLegacyCollisions proves a new agent whose runtime
// name would have produced an already-taken legacy _id can still be created:
// new documents use a random physical identifier.
func TestNewDocumentIDsAvoidLegacyCollisions(t *testing.T) {
	s := newStore(t, testDB(t))
	ctx := context.Background()

	// A legacy tombstone occupies _id "ws:Writer" but has a different name
	// now, so the runtime name "Writer" is free.
	tomb := &agentsv1.Agent{AgentId: "old-writer", Name: "RetiredWriter", WorkspaceId: "ws",
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED}
	spec, err := marshal(tomb)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	doc := configDoc{ID: compositeID("ws", "Writer"), WorkspaceID: "ws", Name: tomb.GetName(), Spec: spec, AgentID: tomb.GetAgentId()}
	if _, err := s.agents.InsertOne(ctx, doc); err != nil {
		t.Fatalf("seed legacy tombstone: %v", err)
	}

	created, err := s.CreateAgent(ctx, "ws", &agentsv1.Agent{AgentId: "writer", Name: "Writer"})
	if err != nil {
		t.Fatalf("CreateAgent with legacy-shaped name: %v", err)
	}
	if created.GetAgentId() != "writer" {
		t.Fatalf("agent_id = %q, want writer", created.GetAgentId())
	}
	if id := physicalID(t, s, "ws", "writer"); strings.Contains(id, ":") {
		t.Fatalf("new document _id = %q, want a random identifier without the legacy composite shape", id)
	}
}

// TestDuplicateRuntimeNameRejectedAtIndex proves the explicit unique
// (workspace_id, name) index preserves the per-workspace runtime-name
// invariant the legacy composite _id used to give for free.
func TestDuplicateRuntimeNameRejectedAtIndex(t *testing.T) {
	s := newStore(t, testDB(t))
	ctx := context.Background()

	if _, err := s.CreateAgent(ctx, "ws", &agentsv1.Agent{AgentId: "writer", Name: "Writer"}); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if _, err := s.CreateAgent(ctx, "ws", &agentsv1.Agent{AgentId: "other", Name: "Writer"}); !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("CreateAgent duplicate name: %v, want ErrAlreadyExists", err)
	}
	// Same name in another workspace is a different scope and must succeed.
	if _, err := s.CreateAgent(ctx, "ws-b", &agentsv1.Agent{AgentId: "writer", Name: "Writer"}); err != nil {
		t.Fatalf("CreateAgent same name other workspace: %v", err)
	}
}
