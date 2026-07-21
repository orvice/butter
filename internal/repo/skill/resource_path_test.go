package skill

import (
	"errors"
	"testing"
)

func TestCleanResourcePath(t *testing.T) {
	valid := []struct {
		name string
		in   string
		want string
	}{
		{"references file", "references/api.md", "references/api.md"},
		{"assets file", "assets/logo.png", "assets/logo.png"},
		{"scripts file", "scripts/run.sh", "scripts/run.sh"},
		{"nested path", "references/deep/dir/notes.md", "references/deep/dir/notes.md"},
		{"redundant segments cleaned", "references/./api.md", "references/api.md"},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CleanResourcePath(tc.in)
			if err != nil {
				t.Fatalf("CleanResourcePath(%q): unexpected error %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("CleanResourcePath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}

	invalid := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"outside spec dirs", "docs/readme.md"},
		{"spec dir itself", "references"},
		{"spec dir trailing slash", "references/"},
		{"leading dotdot", "../references/api.md"},
		{"embedded dotdot staying inside spec dir", "references/dir/../api.md"},
		{"escapes via embedded dotdot", "references/../../etc/passwd"},
		{"dotdot to sibling spec dir", "references/../scripts/run.sh"},
		{"absolute path", "/references/api.md"},
		{"backslash traversal", `references\..\secret`},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := CleanResourcePath(tc.in); !errors.Is(err, ErrInvalidResourcePath) {
				t.Fatalf("CleanResourcePath(%q): expected ErrInvalidResourcePath, got %v", tc.in, err)
			}
		})
	}
}

func TestCleanResourceSubpath(t *testing.T) {
	valid := []struct {
		name string
		in   string
		want string
	}{
		{"bare spec dir", "references", "references"},
		{"trailing slash cleans to bare dir", "assets/", "assets"},
		{"file under dir", "references/api.md", "references/api.md"},
		{"nested dir", "scripts/setup", "scripts/setup"},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CleanResourceSubpath(tc.in)
			if err != nil {
				t.Fatalf("CleanResourceSubpath(%q): unexpected error %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("CleanResourceSubpath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}

	invalid := []string{"", "docs", "../references", "references/..", `references\..\x`, "/references"}
	for _, in := range invalid {
		t.Run(in, func(t *testing.T) {
			if _, err := CleanResourceSubpath(in); !errors.Is(err, ErrInvalidResourcePath) {
				t.Fatalf("CleanResourceSubpath(%q): expected ErrInvalidResourcePath, got %v", in, err)
			}
		})
	}
}
