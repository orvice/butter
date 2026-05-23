package config

import (
	"time"

	"github.com/achetronic/adk-utils-go/plugin/langfuse"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type AppConfig struct {
	Agents           []agentsv1.Agent         `yaml:"agents"`
	Channels         []agentsv1.AgentChannel  `yaml:"channels"`
	ModelProviders   []agentsv1.ModelProvider `yaml:"model_providers"`
	MCPServerConfigs []agentsv1.MCPServer     `yaml:"mcp_server_configs"`
	RemoteAgents     []agentsv1.RemoteAgent   `yaml:"remote_agents"`
	APIToken         string                   `yaml:"apiToken"`
	Auth             AuthConfig               `yaml:"auth"`
	SystemAgentModel string                   `yaml:"system_agent_model"`

	Langfuse langfuse.Config `yaml:"langfuse"`

	MongoURI      string `yaml:"mongo_uri"`
	MongoDB       string `yaml:"mongo_db"`
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`

	HTTP           HTTPConfig       `yaml:"http"`
	Static         StaticConfig     `yaml:"static"`
	Artifact       ArtifactConfig   `yaml:"artifact"`
	AgentFiles     AgentFilesConfig `yaml:"agent_files"`
	MCPOAuth       MCPOAuthConfig   `yaml:"mcp_oauth"`
	StorageBackend string           `yaml:"storage_backend"` // "mongo" (default) or "memory"
	GRPCPort       int              `yaml:"grpc_port"`       // daemon gRPC server port (default 9090)
}

// ArtifactConfig configures the S3-backed ADK artifact service. Artifacts are
// per-(app, user, session) blobs produced or consumed by agents (e.g. a tool
// returning a generated image, a session-scoped attachment).
//
// Storage is delegated to the S3 client registered with butterfly core under
// `store.s3.<S3Bucket>` (see https://butterfly.orz.ee/stores/s3.html). Keep
// this bucket private — unlike static assets, artifacts are not served from
// a CDN and may contain user-specific or model-generated content.
//
// When S3Bucket is empty the artifact service is disabled and ADK falls back
// to its in-process behavior (artifact writes are no-ops / reads fail).
type ArtifactConfig struct {
	// S3Bucket is the key under `store.s3` to use for artifact storage
	// (e.g. "artifacts"). If empty, the artifact service is disabled.
	S3Bucket string `yaml:"s3_bucket"`

	// KeyPrefix is prepended to every object key (e.g. "artifacts/").
	// Optional. Use it to share a bucket with other workloads.
	KeyPrefix string `yaml:"key_prefix"`
}

// Enabled reports whether artifact storage is configured.
func (a ArtifactConfig) Enabled() bool {
	return a.S3Bucket != ""
}

// AgentFilesConfig configures the workspace-scoped text file spaces that
// agents can mount through the built-in agent_files toolset.
type AgentFilesConfig struct {
	// S3Bucket is the key under `store.s3` to use for file contents. Metadata
	// is stored in MongoDB when storage_backend is mongo. If empty, file
	// contents fall back to in-memory storage for local development.
	S3Bucket string `yaml:"s3_bucket"`

	// KeyPrefix is prepended to every object key.
	KeyPrefix string `yaml:"key_prefix"`

	// MaxFileBytes caps one text file write. 0 means use the default 256 KiB.
	MaxFileBytes int64 `yaml:"max_file_bytes"`
}

func (a AgentFilesConfig) EffectiveMaxFileBytes() int64 {
	if a.MaxFileBytes <= 0 {
		return 256 * 1024
	}
	return a.MaxFileBytes
}

// MCPOAuthConfig controls the browser-based OAuth2 flow used for protected
// remote MCP servers. Token material is always persisted server-side.
type MCPOAuthConfig struct {
	// CallbackBaseURL is the externally reachable base URL for this service.
	// The OAuth redirect URI is built as:
	//   <callback_base_url>/api/mcp/oauth/callback
	CallbackBaseURL string `yaml:"callback_base_url"`

	// DashboardBaseURL is used as the fallback redirect target after a backend
	// callback completes and the start request did not provide return_url.
	DashboardBaseURL string `yaml:"dashboard_base_url"`

	// EncryptionKey protects stored token material. It may be raw 32-byte text,
	// hex-encoded bytes, or base64-encoded bytes.
	EncryptionKey string `yaml:"encryption_key"`

	// AllowInsecureHTTP permits non-localhost HTTP OAuth endpoints. It is
	// intended for development only; HTTPS is required otherwise.
	AllowInsecureHTTP bool `yaml:"allow_insecure_http"`
}

type HTTPConfig struct {
	Greeting string `yaml:"greeting"`
}

// StaticConfig configures how user-uploaded static assets (avatars, etc.)
// are stored and served.
//
// Storage is delegated to the S3 client registered with butterfly core under
// `store.s3.<S3Bucket>` (see https://butterfly.orz.ee/stores/s3.html).
// When CDNBaseURL is set, public URLs returned to clients use that base
// instead of pointing at S3 directly — this is how a CDN (e.g. CloudFront,
// Cloudflare R2 custom domain, MinIO behind nginx) is wired in.
type StaticConfig struct {
	// S3Bucket is the key under `store.s3` to use for uploads (e.g. "assets").
	// If empty, upload endpoints are disabled.
	S3Bucket string `yaml:"s3_bucket"`

	// KeyPrefix is prepended to every object key (e.g. "butter/" or "prod/").
	// Optional. Helps when multiple apps share one bucket.
	KeyPrefix string `yaml:"key_prefix"`

	// CDNBaseURL is the public base URL prepended to object keys when
	// returning URLs to clients (e.g. "https://cdn.example.com").
	// If empty, URLs fall back to the configured S3 PublicBaseURL or to
	// "s3://<bucket>/<key>" as a last resort.
	CDNBaseURL string `yaml:"cdn_base_url"`

	// PublicBaseURL is an optional non-CDN public URL for the bucket
	// (e.g. "https://s3.amazonaws.com/my-bucket"). Used only when
	// CDNBaseURL is empty.
	PublicBaseURL string `yaml:"public_base_url"`

	// MaxUploadBytes caps a single upload. 0 means use the default (5 MiB).
	MaxUploadBytes int64 `yaml:"max_upload_bytes"`
}

// EffectiveMaxUploadBytes returns the upload size limit in bytes.
func (s StaticConfig) EffectiveMaxUploadBytes() int64 {
	if s.MaxUploadBytes <= 0 {
		return 5 * 1024 * 1024
	}
	return s.MaxUploadBytes
}

// PublicURL builds a public URL for the given object key honoring
// CDNBaseURL > PublicBaseURL > s3://<bucket>/<key> fallback.
func (s StaticConfig) PublicURL(key string) string {
	base := s.CDNBaseURL
	if base == "" {
		base = s.PublicBaseURL
	}
	if base == "" {
		return "s3://" + s.S3Bucket + "/" + key
	}
	// Trim trailing slash on base and leading slash on key for a clean join.
	for len(base) > 0 && base[len(base)-1] == '/' {
		base = base[:len(base)-1]
	}
	for len(key) > 0 && key[0] == '/' {
		key = key[1:]
	}
	return base + "/" + key
}

type AuthConfig struct {
	InitialAdminUsername string                         `yaml:"initial_admin_username"`
	InitialAdminPassword string                         `yaml:"initial_admin_password"`
	SessionTTL           time.Duration                  `yaml:"session_ttl"`
	OAuthProviders       map[string]OAuthProviderConfig `yaml:"oauth_providers"`
}

// OAuthProviderConfig holds the client credentials and OAuth endpoints for a
// single third-party login provider. Only providers whose ClientID and
// ClientSecret are set are exposed via ListOAuthProviders; the rest are
// effectively disabled.
type OAuthProviderConfig struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
	// DisplayName is shown on the login page (e.g. "GitHub"). Defaults to a
	// titlecase version of the provider key when empty.
	DisplayName string `yaml:"display_name"`
}

// Enabled reports whether the provider has the minimum credentials needed
// to participate in the OAuth flow.
func (c OAuthProviderConfig) Enabled() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

func (c AuthConfig) EffectiveSessionTTL() time.Duration {
	if c.SessionTTL <= 0 {
		return 7 * 24 * time.Hour
	}
	return c.SessionTTL
}

func (c *AppConfig) Print() {}
