package credentials

import (
	"errors"
	"testing"
)

func TestMemoryRoundTrip(t *testing.T) {
	m := NewMemory()
	if _, err := m.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty store err = %v, want ErrNotFound", err)
	}
	value := []byte("token")
	if err := m.Set("k", value); err != nil {
		t.Fatalf("Set: %v", err)
	}
	value[0] = 'X' // the store must not alias the caller's slice
	got, err := m.Get("k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "token" {
		t.Fatalf("got %q, want the value as stored", got)
	}
	got[0] = 'Y'
	again, err := m.Get("k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(again) != "token" {
		t.Fatalf("Get aliased its own storage: %q", again)
	}
}

func TestMemoryDelete(t *testing.T) {
	m := NewMemory()
	if err := m.Delete("absent"); err != nil {
		t.Fatalf("deleting an absent key: %v", err)
	}
	if err := m.Set("k", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete("k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if keys := m.Keys(); len(keys) != 0 {
		t.Fatalf("keys = %v", keys)
	}
}

func TestUnsupportedStore(t *testing.T) {
	var s Store = Unsupported{}
	if _, err := s.Get("k"); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Get err = %v", err)
	}
	if err := s.Set("k", []byte("v")); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Set err = %v", err)
	}
	if err := s.Delete("k"); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Delete err = %v", err)
	}
	if _, ok := s.(Unsupported); !ok {
		t.Fatalf("NewSystem fallback is %T, want Unsupported", s)
	}
}

func TestReadToken(t *testing.T) {
	s := NewMemory()
	if _, err := ReadToken(s, "k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound kept through the wrapper", err)
	}
	if err := s.Set("k", []byte("tok")); err != nil {
		t.Fatal(err)
	}
	tok, err := ReadToken(s, "k")
	if err != nil {
		t.Fatalf("ReadToken: %v", err)
	}
	if tok != "tok" {
		t.Fatalf("token = %q", tok)
	}
}
