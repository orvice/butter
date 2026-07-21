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

// rejectTraversal fails any path with a traversal escape hatch. Any ".."
// segment is rejected outright — even one that would clean to a path still
// inside a spec directory — because this guard is the only traversal
// protection for our custom Source (ADK's FileSystemSource safeguards do not
// apply here). Backslashes are rejected rather than normalized so
// Windows-style traversal cannot slip through path.Clean.
func rejectTraversal(p string) error {
	if p == "" {
		return fmt.Errorf("empty path: %w", ErrInvalidResourcePath)
	}
	if strings.Contains(p, `\`) {
		return fmt.Errorf("path %q contains backslash: %w", p, ErrInvalidResourcePath)
	}
	if slices.Contains(strings.Split(p, "/"), "..") {
		return fmt.Errorf("path %q contains traversal: %w", p, ErrInvalidResourcePath)
	}
	return nil
}

// specDir returns the spec directory a cleaned path lives under (or is), and
// whether it lives under one at all.
func specDir(cleaned string) (string, bool) {
	for _, dir := range resourceDirs {
		if cleaned == dir || strings.HasPrefix(cleaned, dir+"/") {
			return dir, true
		}
	}
	return "", false
}

// CleanResourcePath validates and normalizes a skill resource path. The
// cleaned path must name a file under references/, assets/, or scripts/ (the
// bare directory itself is not a resource).
func CleanResourcePath(p string) (string, error) {
	if err := rejectTraversal(p); err != nil {
		return "", err
	}
	cleaned := path.Clean(p)
	dir, ok := specDir(cleaned)
	if !ok {
		return "", fmt.Errorf("path %q is outside references/, assets/, scripts/: %w", p, ErrInvalidResourcePath)
	}
	if cleaned == dir {
		return "", fmt.Errorf("path %q must name a file under %s/: %w", p, dir, ErrInvalidResourcePath)
	}
	return cleaned, nil
}

// CleanResourceSubpath validates a resource-listing subpath. It shares
// CleanResourcePath's traversal guard, but a bare spec directory
// ("references") is valid here since a listing narrows to a directory rather
// than addressing a single file. Callers handle the root subpath ("" / ".")
// before calling this.
func CleanResourceSubpath(p string) (string, error) {
	if err := rejectTraversal(p); err != nil {
		return "", err
	}
	cleaned := path.Clean(p)
	if _, ok := specDir(cleaned); !ok {
		return "", fmt.Errorf("subpath %q is outside references/, assets/, scripts/: %w", p, ErrInvalidResourcePath)
	}
	return cleaned, nil
}
