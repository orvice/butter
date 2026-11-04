package workstation

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

// SourceIdentity is pinned from the provider probe, never a display name/Host
// record ID or token principal. Tenant separates organizations on one instance.
type SourceIdentity struct {
	Provider string `bson:"provider"`
	Origin   string `bson:"origin"`
	TenantID string `bson:"tenant_id"`
}

func GitHubSource() SourceIdentity {
	return SourceIdentity{Provider: "github", Origin: "https://api.github.com"}
}
func (s SourceIdentity) Canonical() (SourceIdentity, error) {
	s.Provider = strings.ToLower(strings.TrimSpace(s.Provider))
	u, err := url.Parse(s.Origin)
	if err != nil || s.Provider == "" || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.ContainsRune(s.Provider+s.TenantID, '\x00') {
		return SourceIdentity{}, ErrInvalid
	}
	u.Host = strings.ToLower(u.Host)
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		u.Host = u.Hostname()
		if strings.Contains(u.Host, ":") {
			u.Host = "[" + u.Host + "]"
		}
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = ""
	s.Origin = u.String()
	return s, nil
}

// Only the initial GitHub preview wrote source-less records. Preserve that
// projection's namespace while new records always persist their identity.
func (r Record) SourceIdentity() SourceIdentity {
	if r.Source == (SourceIdentity{}) {
		return GitHubSource()
	}
	return r.Source
}
func ItemID(workspaceID string, source SourceIdentity, scopeID, kind, number string) string {
	// Keep the original GitHub byte sequence stable. A nonempty tenant adds its
	// own suffix namespace; no later provider or instance changes this format.
	// Linear and Plane item IDs are globally unique within their tenant. A
	// team/project move changes provenance, not logical identity. GitHub/GitLab
	// use container-local numbers/iids and must keep the stable container ID.
	if source.Provider == "linear" || source.Provider == "plane" {
		scopeID = ""
	}
	key := workspaceID + "\x00" + source.Provider + "\x00" + source.Origin + "\x00" + scopeID + "\x00" + kind + "\x00" + number
	if source.TenantID != "" {
		key += "\x00tenant\x00" + source.TenantID
	}
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
