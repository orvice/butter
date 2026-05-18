// Package service contains business-logic services.
package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"butterfly.orx.me/core/store/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"go.orx.me/apps/butter/internal/config"
)

// ErrUploadDisabled is returned when static upload is not configured.
var ErrUploadDisabled = errors.New("upload: static storage is not configured")

// ErrUploadTooLarge is returned when the upload exceeds the configured limit.
var ErrUploadTooLarge = errors.New("upload: payload exceeds max size")

// ErrUnsupportedContentType is returned when an avatar upload uses a
// content type that the service does not accept.
var ErrUnsupportedContentType = errors.New("upload: unsupported content type")

// avatarAllowedTypes lists the content types accepted by the avatar upload
// endpoint. Keeping this whitelist tight protects the CDN from being used as
// a generic file host.
var avatarAllowedTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// UploadResult describes the outcome of an upload.
type UploadResult struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// UploadService handles object uploads to the configured S3 store and returns
// CDN-aware public URLs.
//
// The service resolves its StaticConfig lazily on every call via the
// provider, because the upload service is constructed during route setup
// (before YAML config has been loaded). Reading the config eagerly at
// construction time would freeze it at the zero value and uploads would be
// permanently disabled even after `static.s3_bucket` is set in config.yaml.
type UploadService struct {
	provider func() config.StaticConfig
}

// NewUploadService builds an UploadService bound to the given static
// configuration snapshot. Mostly useful in tests; production callers should
// prefer NewUploadServiceLazy so the service picks up config loaded after
// route setup.
//
// Returns nil if the configuration disables uploads (S3Bucket empty), to
// preserve historical behavior of "no service → 503".
func NewUploadService(cfg config.StaticConfig) *UploadService {
	if cfg.S3Bucket == "" {
		return nil
	}
	return &UploadService{provider: func() config.StaticConfig { return cfg }}
}

// NewUploadServiceLazy returns a service that reads its StaticConfig from
// the provider on every call. The provider MUST be non-nil and MUST return
// the current config (typically `func() config.StaticConfig { return cfg.Static }`
// against a long-lived *AppConfig). Unlike NewUploadService this never
// returns nil — Enabled() flips to true automatically once the bucket is
// configured.
func NewUploadServiceLazy(provider func() config.StaticConfig) *UploadService {
	if provider == nil {
		return nil
	}
	return &UploadService{provider: provider}
}

func (s *UploadService) cfg() config.StaticConfig {
	if s == nil || s.provider == nil {
		return config.StaticConfig{}
	}
	return s.provider()
}

// Enabled reports whether uploads are configured.
func (s *UploadService) Enabled() bool {
	if s == nil {
		return false
	}
	return s.cfg().S3Bucket != ""
}

// PublicURL returns the CDN-aware URL for the given key. Useful when callers
// already know an object's key (e.g. an avatar URL stored on a user record).
func (s *UploadService) PublicURL(key string) string {
	if s == nil {
		return ""
	}
	return s.cfg().PublicURL(key)
}

// UploadAvatar stores an avatar image for the given owner (user id, agent
// name, etc.) and returns its public URL. The caller is expected to have
// validated that the requester may write avatars for ownerID.
func (s *UploadService) UploadAvatar(ctx context.Context, ownerKind, ownerID string, contentType string, body io.Reader) (*UploadResult, error) {
	if !s.Enabled() {
		return nil, ErrUploadDisabled
	}

	contentType = strings.ToLower(strings.TrimSpace(contentType))
	ext, ok := avatarAllowedTypes[contentType]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedContentType, contentType)
	}

	cfg := s.cfg()
	buf, err := readAtMost(body, cfg.EffectiveMaxUploadBytes())
	if err != nil {
		return nil, err
	}

	key := s.buildAvatarKey(cfg, ownerKind, ownerID, ext)
	return s.putObject(ctx, cfg, key, contentType, buf)
}

// UploadStatic stores an arbitrary static asset under the configured prefix.
// Intended for admin tooling — call sites should authorize first.
func (s *UploadService) UploadStatic(ctx context.Context, name, contentType string, body io.Reader) (*UploadResult, error) {
	if !s.Enabled() {
		return nil, ErrUploadDisabled
	}
	if name == "" {
		return nil, errors.New("upload: name is required")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	cfg := s.cfg()
	buf, err := readAtMost(body, cfg.EffectiveMaxUploadBytes())
	if err != nil {
		return nil, err
	}

	key := joinKey(cfg.KeyPrefix, "static", path.Clean("/"+name))
	return s.putObject(ctx, cfg, key, contentType, buf)
}

func (s *UploadService) putObject(ctx context.Context, cfg config.StaticConfig, key, contentType string, body []byte) (*UploadResult, error) {
	client := s3.GetClient(cfg.S3Bucket)
	if client == nil {
		return nil, fmt.Errorf("upload: s3 client %q is not configured (check store.s3.%s)", cfg.S3Bucket, cfg.S3Bucket)
	}
	bucket := s3.GetBucket(cfg.S3Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("upload: s3 bucket name for %q is empty", cfg.S3Bucket)
	}

	_, err := client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("upload: put object %q: %w", key, err)
	}

	return &UploadResult{
		Key:         key,
		URL:         cfg.PublicURL(key),
		ContentType: contentType,
		Size:        int64(len(body)),
	}, nil
}

func (s *UploadService) buildAvatarKey(cfg config.StaticConfig, ownerKind, ownerID, ext string) string {
	ownerKind = sanitizePathComponent(ownerKind)
	if ownerKind == "" {
		ownerKind = "unknown"
	}
	ownerID = sanitizePathComponent(ownerID)
	if ownerID == "" {
		ownerID = "anonymous"
	}
	// Random suffix forces a new URL on each upload so CDN caches refresh.
	suffix := randomSuffix(8)
	stamp := time.Now().UTC().Format("20060102150405")
	name := fmt.Sprintf("%s-%s%s", stamp, suffix, ext)
	return joinKey(cfg.KeyPrefix, "avatars", ownerKind, ownerID, name)
}

func joinKey(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, "/")
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, "/")
}

func sanitizePathComponent(in string) string {
	in = strings.TrimSpace(in)
	var b strings.Builder
	for _, r := range in {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func randomSuffix(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buf)
}

func readAtMost(r io.Reader, max int64) ([]byte, error) {
	limited := io.LimitReader(r, max+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("upload: read body: %w", err)
	}
	if int64(len(buf)) > max {
		return nil, fmt.Errorf("%w: limit=%d bytes", ErrUploadTooLarge, max)
	}
	return buf, nil
}
