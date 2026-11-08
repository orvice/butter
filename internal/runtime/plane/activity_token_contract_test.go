package plane

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestActivityHTTPClientForceRefreshOnlyOnce(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	var refresh int32
	c := &ActivityHTTPClient{BaseURL: srv.URL, HTTPClient: srv.Client(), Token: func(context.Context, string, string, string) (string, error) { return "old", nil }, ForceRefresh: func(context.Context, string, string, string) (string, error) {
		atomic.AddInt32(&refresh, 1)
		return "new", nil
	}}
	_, err := c.CreateActivity(context.Background(), ActivityRequest{WorkspaceID: "w", AppID: "a", InstallationID: "i", WorkspaceSlug: "slug", RunID: "run", Type: ActivityResponse, Body: "safe"})
	if err == nil || calls != 2 || refresh != 1 {
		t.Fatalf("calls=%d refresh=%d err=%v", calls, refresh, err)
	}
	if strings.Contains(err.Error(), "old") || strings.Contains(err.Error(), "new") || strings.Contains(err.Error(), "safe") {
		t.Fatalf("unsafe error: %v", err)
	}
}

func TestActivityHTTPClientDoesNotRefreshPermanentFailure(t *testing.T) {
	var refresh int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer srv.Close()
	c := &ActivityHTTPClient{BaseURL: srv.URL, HTTPClient: srv.Client(), Token: func(context.Context, string, string, string) (string, error) { return "secret", nil }, ForceRefresh: func(context.Context, string, string, string) (string, error) {
		atomic.AddInt32(&refresh, 1)
		return "", nil
	}}
	_, err := c.CreateActivity(context.Background(), ActivityRequest{WorkspaceID: "w", AppID: "a", InstallationID: "i", WorkspaceSlug: "slug", RunID: "run", Type: ActivityResponse, Body: "payload"})
	if err == nil || refresh != 0 || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "payload") {
		t.Fatalf("refresh=%d err=%v", refresh, err)
	}
}
