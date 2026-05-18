package service

import (
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/config"
)

func TestNewUploadServiceDisabledWhenBucketEmpty(t *testing.T) {
	t.Parallel()
	if got := NewUploadService(config.StaticConfig{}); got != nil {
		t.Fatalf("expected nil service when S3Bucket empty, got %#v", got)
	}
}

func TestUploadServiceEnabled(t *testing.T) {
	t.Parallel()
	s := NewUploadService(config.StaticConfig{S3Bucket: "assets"})
	if !s.Enabled() {
		t.Fatal("expected service to be enabled")
	}
}

func TestStaticConfigPublicURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		cfg  config.StaticConfig
		key  string
		want string
	}{
		{
			name: "cdn wins over public base",
			cfg:  config.StaticConfig{S3Bucket: "b", CDNBaseURL: "https://cdn.example.com/", PublicBaseURL: "https://s3.example.com"},
			key:  "/avatars/u/1.png",
			want: "https://cdn.example.com/avatars/u/1.png",
		},
		{
			name: "public base when no cdn",
			cfg:  config.StaticConfig{S3Bucket: "b", PublicBaseURL: "https://s3.example.com/b"},
			key:  "static/site.css",
			want: "https://s3.example.com/b/static/site.css",
		},
		{
			name: "fallback to s3 scheme",
			cfg:  config.StaticConfig{S3Bucket: "my-bucket"},
			key:  "foo/bar",
			want: "s3://my-bucket/foo/bar",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.cfg.PublicURL(tc.key)
			if got != tc.want {
				t.Fatalf("PublicURL(%q) = %q, want %q", tc.key, got, tc.want)
			}
		})
	}
}

func TestBuildAvatarKeyShape(t *testing.T) {
	t.Parallel()
	s := NewUploadService(config.StaticConfig{S3Bucket: "assets", KeyPrefix: "butter"})
	key := s.buildAvatarKey("user", "u-123", ".png")
	if !strings.HasPrefix(key, "butter/avatars/user/u-123/") {
		t.Fatalf("unexpected key prefix: %q", key)
	}
	if !strings.HasSuffix(key, ".png") {
		t.Fatalf("expected .png suffix, got %q", key)
	}
}

func TestSanitizePathComponent(t *testing.T) {
	t.Parallel()
	if got := sanitizePathComponent("foo/bar baz"); got != "foo_bar_baz" {
		t.Fatalf("unexpected sanitize result: %q", got)
	}
}

func TestReadAtMostRejectsTooLarge(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(strings.Repeat("a", 11))
	if _, err := readAtMost(r, 10); err == nil {
		t.Fatal("expected too-large error")
	}
}

func TestReadAtMostAcceptsAtLimit(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(strings.Repeat("a", 10))
	buf, err := readAtMost(r, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buf) != 10 {
		t.Fatalf("expected 10 bytes, got %d", len(buf))
	}
}
