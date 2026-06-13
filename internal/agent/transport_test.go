package agent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResponseHeaderTimeoutTransport(t *testing.T) {
	client := &http.Client{Transport: &responseHeaderTimeoutTransport{
		base:    http.DefaultTransport,
		timeout: 100 * time.Millisecond,
	}}

	t.Run("server never sends headers", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer srv.Close()

		_, err := client.Get(srv.URL)
		if err == nil || !strings.Contains(err.Error(), "mcp request exceeded") {
			t.Fatalf("expected timeout error, got %v", err)
		}
	})

	t.Run("standalone listener (GET) body is unaffected after headers", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			time.Sleep(300 * time.Millisecond)
			_, _ = w.Write([]byte("done"))
		}))
		defer srv.Close()

		resp, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(body) != "done" {
			t.Fatalf("body = %q, want %q", body, "done")
		}
	})

	t.Run("message POST body that stalls after headers times out", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			<-r.Context().Done() // never finish the body
		}))
		defer srv.Close()

		resp, err := client.Post(srv.URL, "application/json", strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		_, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err == nil {
			t.Fatal("expected body read to fail due to timeout, got nil")
		}
	})

	t.Run("message POST that completes quickly succeeds", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}))
		defer srv.Close()

		resp, err := client.Post(srv.URL, "application/json", strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(body) != "ok" {
			t.Fatalf("body = %q, want %q", body, "ok")
		}
	})
}
