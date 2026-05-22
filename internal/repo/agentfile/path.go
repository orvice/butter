package agentfile

import (
	"path"
	"strings"
)

// NormalizePath returns a canonical absolute path inside an agent file space.
func NormalizePath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", ErrInvalidPath
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	clean := path.Clean(p)
	if clean == "." || clean == "/" || strings.Contains(clean, "\x00") {
		return "", ErrInvalidPath
	}
	if strings.HasPrefix(clean, "/../") || clean == "/.." {
		return "", ErrInvalidPath
	}
	return clean, nil
}

// NormalizePrefix returns a canonical absolute prefix. Empty means all files.
func NormalizePrefix(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return "", nil
	}
	return NormalizePath(p)
}
