package agentfile

import "testing"

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"notes.md", "/notes.md", false},
		{"/notes/../todo.md", "/todo.md", false},
		{"", "", true},
		{"/", "", true},
		{"\x00", "", true},
	}

	for _, tc := range cases {
		got, err := NormalizePath(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("NormalizePath(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizePath(%q) unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
