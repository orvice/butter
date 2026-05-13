package repo

import "testing"

func TestNewHealthRepository(t *testing.T) {
	r := NewHealthRepository()
	if r == nil {
		t.Fatal("NewHealthRepository returned nil")
	}
}

func TestHealthRepository_ServiceName(t *testing.T) {
	r := NewHealthRepository()
	if got, want := r.ServiceName(), "butter"; got != want {
		t.Fatalf("ServiceName() = %q, want %q", got, want)
	}
}
