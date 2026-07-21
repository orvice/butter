package skill

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"
)

// ErrInvalidResourcePath marks resource paths that fall outside the spec
// directories or attempt traversal. Callers map it to InvalidArgument.
var ErrInvalidResourcePath = errors.New("invalid resource path")

// resourceDirs are the agentskills.io spec directories a skill resource may
// live under.
var resourceDirs = []string{"references", "assets", "scripts"}

// CleanResourcePath validates and normalizes a skill resource path. The
// cleaned path must name a file under references/, assets/, or scripts/.
// Any ".." segment is rejected outright — even one that would clean to a
// path still inside a spec directory — because this guard is the only
// traversal protection for our custom Source (ADK's FileSystemSource
// safeguards do not apply here). Backslashes are rejected rather than
// normalized so Windows-style traversal cannot slip through path.Clean.
func CleanResourcePath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("empty path: %w", ErrInvalidResourcePath)
	}
	if strings.Contains(p, `\`) {
		return "", fmt.Errorf("path %q contains backslash: %w", p, ErrInvalidResourcePath)
	}
	if slices.Contains(strings.Split(p, "/"), "..") {
		return "", fmt.Errorf("path %q contains traversal: %w", p, ErrInvalidResourcePath)
	}
	cleaned := path.Clean(p)
	for _, dir := range resourceDirs {
		if strings.HasPrefix(cleaned, dir+"/") && len(cleaned) > len(dir)+1 {
			return cleaned, nil
		}
	}
	return "", fmt.Errorf("path %q is outside references/, assets/, scripts/: %w", p, ErrInvalidResourcePath)
}
