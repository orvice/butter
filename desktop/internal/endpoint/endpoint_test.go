package endpoint

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		key  bool
	}{
		{name: "host gets https", in: "butter.example.com", want: "https://butter.example.com"},
		{name: "keeps http for local", in: "http://localhost:8787", want: "http://localhost:8787"},
		{name: "trailing slash dropped", in: "https://butter.example.com/", want: "https://butter.example.com"},
		{name: "path prefix kept", in: "https://example.com/butter/", want: "https://example.com/butter"},
		{name: "surrounding space trimmed", in: "  butter.example.com  ", want: "https://butter.example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tc.in, err)
			}
			if got.String() != tc.want {
				t.Fatalf("Parse(%q) = %q, want %q", tc.in, got.String(), tc.want)
			}
			if !got.Valid() {
				t.Fatalf("Parse(%q) is not valid", tc.in)
			}
		})
	}
}

func TestParseRejects(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"ftp://butter.example.com",
		"https://",
		"https://user:pass@butter.example.com",
		"https://butter.example.com/?a=1",
		"https://butter.example.com/#x",
	}
	for _, in := range cases {
		if _, err := Parse(in); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Parse(%q) error = %v, want ErrInvalid", in, err)
		}
	}
}

func TestAPIBase(t *testing.T) {
	s, err := Parse("butter.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := s.APIBase(), "https://butter.example.com/api"; got != want {
		t.Fatalf("APIBase() = %q, want %q", got, want)
	}
}

func TestKeyFollowsNormalizedAddress(t *testing.T) {
	a, err := Parse("butter.example.com")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Parse("https://butter.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	c, err := Parse("https://other.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if a.Key() == "" {
		t.Fatal("empty key")
	}
	if a.Key() != b.Key() {
		t.Fatalf("same address has different keys: %q vs %q", a.Key(), b.Key())
	}
	if a.Key() == c.Key() {
		t.Fatal("different addresses share a key")
	}
}
