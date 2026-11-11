// Package prefs stores the window's non-secret preferences: the server
// address, the last username and the chosen workspace. Session tokens live
// in the Keychain, never here.
package prefs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Data is the persisted state. It has no field for a password or a token,
// by design.
type Data struct {
	Server      string `json:"server,omitempty"`
	Username    string `json:"username,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

// DefaultPath is <user config dir>/butter/desktop.json.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("prefs: user config directory: %w", err)
	}
	return filepath.Join(dir, "butter", "desktop.json"), nil
}

// Load reads path. A missing file is not an error: it returns the zero
// Data, as the first launch wants.
func Load(path string) (Data, error) {
	var d Data
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Data{}, nil
	}
	if err != nil {
		return Data{}, fmt.Errorf("prefs: read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return Data{}, fmt.Errorf("prefs: parse %s: %w", path, err)
	}
	return d, nil
}

// Save writes path atomically, readable by the user only.
func Save(path string, d Data) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("prefs: create %s: %w", filepath.Dir(path), err)
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("prefs: encode: %w", err)
	}
	b = append(b, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".desktop-*.json")
	if err != nil {
		return fmt.Errorf("prefs: create temporary file: %w", err)
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("prefs: set permissions: %w", err)
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("prefs: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("prefs: sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("prefs: close: %w", err)
	}
	if err := os.Rename(name, path); err != nil {
		return fmt.Errorf("prefs: replace %s: %w", path, err)
	}
	return nil
}
