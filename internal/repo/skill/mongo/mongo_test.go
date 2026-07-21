package mongo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	skillmongo "go.orx.me/apps/butter/internal/repo/skill/mongo"
	"go.orx.me/apps/butter/internal/repo/skill/repotest"
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

	db := client.Database(fmt.Sprintf("butter_skill_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func newStore(t *testing.T, db *mongo.Database, content skillrepo.ContentStore) *skillmongo.Store {
	t.Helper()
	store := skillmongo.New(db, content)
	if err := store.EnsureIndexes(context.Background()); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	return store
}

// spyContentStore counts calls so tests can assert the runtime hot path
// (List / Get) never touches content storage (issue #153).
type spyContentStore struct {
	inner   skillrepo.ContentStore
	puts    atomic.Int64
	gets    atomic.Int64
	deletes atomic.Int64
}

func (s *spyContentStore) Put(ctx context.Context, key, content string) error {
	s.puts.Add(1)
	return s.inner.Put(ctx, key, content)
}

func (s *spyContentStore) Get(ctx context.Context, key string) (string, error) {
	s.gets.Add(1)
	return s.inner.Get(ctx, key)
}

func (s *spyContentStore) Delete(ctx context.Context, keys []string) error {
	s.deletes.Add(1)
	return s.inner.Delete(ctx, keys)
}

func TestMongoRepositoryConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) skillrepo.Repository {
		return newStore(t, testDB(t), skillrepo.NewMemoryContentStore())
	})
}

const testSkillMD = "---\nname: pdf-report\ndescription: test skill\n---\n# body\n"

func testSkill() *agentsv1.Skill {
	return &agentsv1.Skill{Name: "pdf-report", Description: "test skill"}
}

func TestMongoRepositorySurvivesRestart(t *testing.T) {
	db := testDB(t)
	content := skillrepo.NewMemoryContentStore()
	first := newStore(t, db, content)

	ctx := context.Background()
	if _, err := first.Create(ctx, "ws-a", testSkill(), testSkillMD); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// A fresh Store over the same database models a process restart. The
	// in-memory content store survives here because restart durability of
	// content is S3's concern; metadata durability is what Mongo adds.
	second := newStore(t, db, content)
	if _, err := second.Get(ctx, "ws-a", "pdf-report"); err != nil {
		t.Fatalf("Get after restart: %v", err)
	}
	skills, err := second.List(ctx, "ws-a")
	if err != nil || len(skills) != 1 {
		t.Fatalf("List after restart: %v (n=%d)", err, len(skills))
	}
	md, err := second.GetSkillMD(ctx, "ws-a", "pdf-report")
	if err != nil || md != testSkillMD {
		t.Fatalf("GetSkillMD after restart: %v (md=%q)", err, md)
	}
}

func TestMongoRepositoryHotPathDoesNotTouchContentStore(t *testing.T) {
	spy := &spyContentStore{inner: skillrepo.NewMemoryContentStore()}
	store := newStore(t, testDB(t), spy)

	ctx := context.Background()
	if _, err := store.Create(ctx, "ws-a", testSkill(), testSkillMD); err != nil {
		t.Fatalf("Create: %v", err)
	}
	spy.gets.Store(0)

	if _, err := store.List(ctx, "ws-a"); err != nil {
		t.Fatalf("List: %v", err)
	}
	if _, err := store.Get(ctx, "ws-a", "pdf-report"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if n := spy.gets.Load(); n != 0 {
		t.Fatalf("List/Get touched the content store %d times", n)
	}

	if _, err := store.GetSkillMD(ctx, "ws-a", "pdf-report"); err != nil {
		t.Fatalf("GetSkillMD: %v", err)
	}
	if n := spy.gets.Load(); n != 1 {
		t.Fatalf("GetSkillMD should read content exactly once, got %d", n)
	}
}

func TestMongoRepositoryDeleteRemovesStoredContent(t *testing.T) {
	content := skillrepo.NewMemoryContentStore()
	store := newStore(t, testDB(t), content)

	ctx := context.Background()
	if _, err := store.Create(ctx, "ws-a", testSkill(), testSkillMD); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := store.Delete(ctx, "ws-a", "pdf-report"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := content.Get(ctx, skillrepo.ContentKey("ws-a", "pdf-report"))
	if !errors.Is(err, skillrepo.ErrNotFound) {
		t.Fatalf("expected stored content removed, got %v", err)
	}
}

// failingDeleteContentStore simulates S3 being unavailable during delete.
type failingDeleteContentStore struct {
	skillrepo.ContentStore
}

func (f *failingDeleteContentStore) Delete(context.Context, []string) error {
	return errors.New("s3 unavailable")
}

// Metadata is the source of truth: once the Mongo document is gone the skill
// is deleted, even if removing the stored body fails. The orphaned object is
// logged, not surfaced as a delete failure (a retry would just get NotFound).
func TestMongoRepositoryDeleteSucceedsWhenContentDeleteFails(t *testing.T) {
	content := &failingDeleteContentStore{ContentStore: skillrepo.NewMemoryContentStore()}
	store := newStore(t, testDB(t), content)

	ctx := context.Background()
	if _, err := store.Create(ctx, "ws-a", testSkill(), testSkillMD); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := store.Delete(ctx, "ws-a", "pdf-report"); err != nil {
		t.Fatalf("Delete should tolerate content-store failure, got %v", err)
	}
	if _, err := store.Get(ctx, "ws-a", "pdf-report"); !errors.Is(err, skillrepo.ErrNotFound) {
		t.Fatalf("expected skill gone after delete, got %v", err)
	}
}

func TestMongoRepositoryDuplicateCreateKeepsExistingContent(t *testing.T) {
	store := newStore(t, testDB(t), skillrepo.NewMemoryContentStore())

	ctx := context.Background()
	if _, err := store.Create(ctx, "ws-a", testSkill(), testSkillMD); err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, err := store.Create(ctx, "ws-a", testSkill(), "---\nname: pdf-report\ndescription: clobber\n---\nother\n")
	if !errors.Is(err, skillrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
	md, err := store.GetSkillMD(ctx, "ws-a", "pdf-report")
	if err != nil {
		t.Fatalf("GetSkillMD: %v", err)
	}
	if md != testSkillMD {
		t.Fatalf("duplicate create clobbered stored content:\n%s", md)
	}
}

func TestMongoResourceListServedFromPathIndex(t *testing.T) {
	spy := &spyContentStore{inner: skillrepo.NewMemoryContentStore()}
	store := newStore(t, testDB(t), spy)

	ctx := context.Background()
	if _, err := store.Create(ctx, "ws-a", testSkill(), testSkillMD); err != nil {
		t.Fatalf("Create: %v", err)
	}
	for _, path := range []string{"references/api.md", "assets/logo.png"} {
		if _, err := store.PutResource(ctx, "ws-a", "pdf-report", &agentsv1.SkillResource{Path: path}, []byte("data")); err != nil {
			t.Fatalf("PutResource %s: %v", path, err)
		}
	}
	spy.gets.Store(0)

	resources, err := store.ListResources(ctx, "ws-a", "pdf-report")
	if err != nil || len(resources) != 2 {
		t.Fatalf("ListResources: %v (n=%d)", err, len(resources))
	}
	if n := spy.gets.Load(); n != 0 {
		t.Fatalf("ListResources touched the content store %d times; must be served from the Mongo path index", n)
	}

	if _, _, err := store.GetResource(ctx, "ws-a", "pdf-report", "references/api.md"); err != nil {
		t.Fatalf("GetResource: %v", err)
	}
	if n := spy.gets.Load(); n != 1 {
		t.Fatalf("GetResource should read content exactly once, got %d", n)
	}
}

func TestMongoDeleteSkillBatchesResourceContentDeletes(t *testing.T) {
	spy := &spyContentStore{inner: skillrepo.NewMemoryContentStore()}
	store := newStore(t, testDB(t), spy)

	ctx := context.Background()
	if _, err := store.Create(ctx, "ws-a", testSkill(), testSkillMD); err != nil {
		t.Fatalf("Create: %v", err)
	}
	for _, path := range []string{"references/a.md", "references/b.md", "scripts/c.sh"} {
		if _, err := store.PutResource(ctx, "ws-a", "pdf-report", &agentsv1.SkillResource{Path: path}, []byte("data")); err != nil {
			t.Fatalf("PutResource %s: %v", path, err)
		}
	}
	spy.deletes.Store(0)

	if err := store.Delete(ctx, "ws-a", "pdf-report"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if n := spy.deletes.Load(); n != 1 {
		t.Fatalf("DeleteSkill should batch content deletion into one call, got %d", n)
	}
}
