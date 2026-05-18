package adkutils

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/adk/artifact"
	"google.golang.org/genai"
)

func TestS3ArtifactService(t *testing.T) {
	srv := newS3ArtifactServiceWithClient("bucket", newFakeS3Client())
	ctx := t.Context()

	testData := []struct {
		fileName string
		version  int64
		part     *genai.Part
	}{
		{"file1", 1, genai.NewPartFromBytes([]byte("file v1"), "text/plain")},
		{"file1", 2, genai.NewPartFromBytes([]byte("file v2"), "text/plain")},
		{"file1", 3, genai.NewPartFromBytes([]byte("file v3"), "text/plain")},
		{"file2", 1, genai.NewPartFromBytes([]byte("file2"), "application/octet-stream")},
		{"file3", 1, genai.NewPartFromText("file3")},
		{"user:file4", 1, genai.NewPartFromBytes([]byte("user file"), "text/plain")},
	}

	for _, data := range testData {
		got, err := srv.Save(ctx, &artifact.SaveRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: data.fileName, Part: data.part,
		})
		if err != nil {
			t.Fatalf("Save(%q) failed: %v", data.fileName, err)
		}
		if got.Version != data.version {
			t.Fatalf("Save(%q) version = %d, want %d", data.fileName, got.Version, data.version)
		}
	}

	t.Run("LoadLatest", func(t *testing.T) {
		got, err := srv.Load(ctx, &artifact.LoadRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: "file1",
		})
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}
		want := genai.NewPartFromBytes([]byte("file v3"), "text/plain")
		if diff := cmp.Diff(want, got.Part); diff != "" {
			t.Fatalf("Load() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("List", func(t *testing.T) {
		got, err := srv.List(ctx, &artifact.ListRequest{
			AppName: "app", UserID: "user", SessionID: "session",
		})
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}
		want := []string{"file1", "file2", "file3", "user:file4"}
		if diff := cmp.Diff(want, got.FileNames); diff != "" {
			t.Fatalf("List() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Versions", func(t *testing.T) {
		got, err := srv.Versions(ctx, &artifact.VersionsRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: "file1",
		})
		if err != nil {
			t.Fatalf("Versions() failed: %v", err)
		}
		want := []int64{1, 2, 3}
		if diff := cmp.Diff(want, got.Versions); diff != "" {
			t.Fatalf("Versions() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("GetArtifactVersion", func(t *testing.T) {
		got, err := srv.GetArtifactVersion(ctx, &artifact.GetArtifactVersionRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: "file2",
		})
		if err != nil {
			t.Fatalf("GetArtifactVersion() failed: %v", err)
		}
		want := &artifact.ArtifactVersion{
			Version:        1,
			CanonicalURI:   "s3://bucket/app/user/session/file2/1",
			CustomMetadata: map[string]any{},
			CreateTime:     got.ArtifactVersion.CreateTime,
			MimeType:       "application/octet-stream",
		}
		if diff := cmp.Diff(want, got.ArtifactVersion); diff != "" {
			t.Fatalf("GetArtifactVersion() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("DeleteSpecificVersion", func(t *testing.T) {
		if err := srv.Delete(ctx, &artifact.DeleteRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: "file1", Version: 3,
		}); err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}
		got, err := srv.Load(ctx, &artifact.LoadRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: "file1",
		})
		if err != nil {
			t.Fatalf("Load() after delete failed: %v", err)
		}
		want := genai.NewPartFromBytes([]byte("file v2"), "text/plain")
		if diff := cmp.Diff(want, got.Part); diff != "" {
			t.Fatalf("Load() after delete mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("DeleteAllVersions", func(t *testing.T) {
		if err := srv.Delete(ctx, &artifact.DeleteRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: "file1",
		}); err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}
		got, err := srv.Load(ctx, &artifact.LoadRequest{
			AppName: "app", UserID: "user", SessionID: "session", FileName: "file1",
		})
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Load() after delete = (%v, %v), want fs.ErrNotExist", got, err)
		}
	})
}

func TestS3ArtifactServiceExplicitVersion(t *testing.T) {
	srv := newS3ArtifactServiceWithClient("bucket", newFakeS3Client())
	ctx := t.Context()

	got, err := srv.Save(ctx, &artifact.SaveRequest{
		AppName: "app", UserID: "user", SessionID: "session", FileName: "file",
		Version: 7,
		Part:    genai.NewPartFromText("explicit"),
	})
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	if got.Version != 7 {
		t.Fatalf("Save() version = %d, want 7", got.Version)
	}

	loaded, err := srv.Load(ctx, &artifact.LoadRequest{
		AppName: "app", UserID: "user", SessionID: "session", FileName: "file", Version: 7,
	})
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	want := genai.NewPartFromBytes([]byte("explicit"), defaultTextContentType)
	if diff := cmp.Diff(want, loaded.Part); diff != "" {
		t.Fatalf("Load() mismatch (-want +got):\n%s", diff)
	}
}

type fakeS3Object struct {
	body        []byte
	contentType string
	metadata    map[string]string
	modified    time.Time
}

type fakeS3Client struct {
	mu      sync.Mutex
	objects map[string]fakeS3Object
}

func newFakeS3Client() *fakeS3Client {
	return &fakeS3Client{objects: map[string]fakeS3Object{}}
}

func (c *fakeS3Client) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	body, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	c.objects[aws.ToString(in.Key)] = fakeS3Object{
		body:        body,
		contentType: aws.ToString(in.ContentType),
		metadata:    mapsClone(in.Metadata),
		modified:    time.Now(),
	}
	return &s3.PutObjectOutput{}, nil
}

func (c *fakeS3Client) GetObject(_ context.Context, in *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	obj, ok := c.objects[aws.ToString(in.Key)]
	if !ok {
		return nil, &types.NoSuchKey{}
	}
	return &s3.GetObjectOutput{
		Body:        io.NopCloser(bytes.NewReader(obj.body)),
		ContentType: aws.String(obj.contentType),
	}, nil
}

func (c *fakeS3Client) HeadObject(_ context.Context, in *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	obj, ok := c.objects[aws.ToString(in.Key)]
	if !ok {
		return nil, &types.NotFound{}
	}
	return &s3.HeadObjectOutput{
		ContentType:  aws.String(obj.contentType),
		LastModified: aws.Time(obj.modified),
		Metadata:     mapsClone(obj.metadata),
	}, nil
}

func (c *fakeS3Client) DeleteObject(_ context.Context, in *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.objects, aws.ToString(in.Key))
	return &s3.DeleteObjectOutput{}, nil
}

func (c *fakeS3Client) DeleteObjects(_ context.Context, in *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, obj := range in.Delete.Objects {
		delete(c.objects, aws.ToString(obj.Key))
	}
	return &s3.DeleteObjectsOutput{}, nil
}

func (c *fakeS3Client) ListObjectsV2(_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0)
	for key := range c.objects {
		if strings.HasPrefix(key, aws.ToString(in.Prefix)) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	start := 0
	if in.ContinuationToken != nil {
		start = sort.SearchStrings(keys, aws.ToString(in.ContinuationToken))
	}
	end := min(start+2, len(keys))

	contents := make([]types.Object, 0, end-start)
	for _, key := range keys[start:end] {
		contents = append(contents, types.Object{Key: aws.String(key)})
	}

	out := &s3.ListObjectsV2Output{Contents: contents}
	if end < len(keys) {
		out.IsTruncated = aws.Bool(true)
		out.NextContinuationToken = aws.String(keys[end])
	}
	return out, nil
}

func mapsClone(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

var _ s3ArtifactClient = (*fakeS3Client)(nil)
