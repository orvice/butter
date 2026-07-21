package skill_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
)

type fakeS3Client struct {
	objects      map[string]string
	buckets      map[string]struct{}
	contentTypes map[string]string
}

func newFakeS3Client() *fakeS3Client {
	return &fakeS3Client{
		objects:      make(map[string]string),
		buckets:      make(map[string]struct{}),
		contentTypes: make(map[string]string),
	}
}

func (f *fakeS3Client) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	data, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	f.buckets[aws.ToString(in.Bucket)] = struct{}{}
	f.objects[aws.ToString(in.Key)] = string(data)
	f.contentTypes[aws.ToString(in.Key)] = aws.ToString(in.ContentType)
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3Client) GetObject(_ context.Context, in *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	content, ok := f.objects[aws.ToString(in.Key)]
	if !ok {
		return nil, &types.NoSuchKey{}
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(content))}, nil
}

func (f *fakeS3Client) DeleteObjects(_ context.Context, in *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	for _, obj := range in.Delete.Objects {
		delete(f.objects, aws.ToString(obj.Key))
	}
	return &s3.DeleteObjectsOutput{}, nil
}

// Both ContentStore implementations must report a missing key as
// ErrNotFound so callers see one error contract regardless of backend.
func TestS3ContentStoreGetMissingIsNotFound(t *testing.T) {
	store := skillrepo.NewS3ContentStore("butter-bucket", newFakeS3Client(), "skills")

	_, err := store.Get(t.Context(), skillrepo.ContentKey("ws-a", "absent"))
	if !errors.Is(err, skillrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestS3ContentStorePutUsesDocumentedKeyShape(t *testing.T) {
	client := newFakeS3Client()
	store := skillrepo.NewS3ContentStore("butter-bucket", client, "skills")

	key := skillrepo.ContentKey("ws-a", "pdf-report")
	if err := store.Put(t.Context(), key, "# PDF Report\n"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	// Issue #153: objects land under <key_prefix>/<workspace_id>/<skill_name>/SKILL.md.
	got, ok := client.objects["skills/ws-a/pdf-report/SKILL.md"]
	if !ok {
		t.Fatalf("object not stored under documented key, stored keys: %v", client.objects)
	}
	if got != "# PDF Report\n" {
		t.Fatalf("unexpected stored content %q", got)
	}
	if _, ok := client.buckets["butter-bucket"]; !ok {
		t.Fatalf("object not stored in configured bucket")
	}
}

func TestS3ContentStoreGetRoundTrip(t *testing.T) {
	client := newFakeS3Client()
	store := skillrepo.NewS3ContentStore("butter-bucket", client, "skills/")

	key := skillrepo.ContentKey("ws-a", "pdf-report")
	if err := store.Put(t.Context(), key, "body"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := store.Get(t.Context(), key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "body" {
		t.Fatalf("content did not round-trip, got %q", got)
	}
}

func TestS3ContentStoreDeleteRemovesObjects(t *testing.T) {
	client := newFakeS3Client()
	store := skillrepo.NewS3ContentStore("butter-bucket", client, "skills")

	key := skillrepo.ContentKey("ws-a", "pdf-report")
	if err := store.Put(t.Context(), key, "body"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := store.Delete(t.Context(), []string{key}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(client.objects) != 0 {
		t.Fatalf("expected no objects after delete, got %v", client.objects)
	}
}

func TestS3ContentStoreWithoutPrefixUsesBareKey(t *testing.T) {
	client := newFakeS3Client()
	store := skillrepo.NewS3ContentStore("butter-bucket", client, "")

	key := skillrepo.ContentKey("ws-a", "pdf-report")
	if err := store.Put(t.Context(), key, "body"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, ok := client.objects["ws-a/pdf-report/SKILL.md"]; !ok {
		t.Fatalf("expected bare key without prefix, stored keys: %v", client.objects)
	}
}

// Resource objects are arbitrary binary files; only SKILL.md keys should be
// tagged as markdown (issue #154).
func TestS3ContentStoreContentTypeByKeyKind(t *testing.T) {
	client := newFakeS3Client()
	store := skillrepo.NewS3ContentStore("skill-bucket", client, "skills")

	if err := store.Put(context.Background(), skillrepo.ContentKey("ws-a", "pdf-report"), "# doc"); err != nil {
		t.Fatalf("Put SKILL.md: %v", err)
	}
	if err := store.Put(context.Background(), skillrepo.ResourceContentKey("ws-a", "pdf-report", "assets/logo.png"), "\x89PNG"); err != nil {
		t.Fatalf("Put resource: %v", err)
	}

	if got := client.contentTypes["skills/ws-a/pdf-report/SKILL.md"]; got != "text/markdown; charset=utf-8" {
		t.Fatalf("SKILL.md content type = %q", got)
	}
	if got := client.contentTypes["skills/ws-a/pdf-report/assets/logo.png"]; got != "application/octet-stream" {
		t.Fatalf("resource content type = %q", got)
	}
}
