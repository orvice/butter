package agent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func timeoutClient(transport agentsv1.MCPServerTransport) *http.Client {
	return &http.Client{Transport: &responseHeaderTimeoutTransport{
		base:      http.DefaultTransport,
		timeout:   100 * time.Millisecond,
		transport: transport,
	}}
}

func TestResponseHeaderTimeoutTransport(t *testing.T) {
	t.Run("server never sends headers times out", func(t *testing.T) {
		client := timeoutClient(agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer srv.Close()

		_, err := client.Get(srv.URL)
		if err == nil || !strings.Contains(err.Error(), "mcp request exceeded") {
			t.Fatalf("expected timeout error, got %v", err)
		}
	})

	t.Run("streamable HTTP standalone listener (GET) unaffected after headers", func(t *testing.T) {
		client := timeoutClient(agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			time.Sleep(300 * time.Millisecond) // idle listener, no data yet
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

	t.Run("legacy SSE handshake (GET) that never sends endpoint times out", func(t *testing.T) {
		client := timeoutClient(agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_SSE)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			<-r.Context().Done() // headers sent, but no endpoint event
		}))
		defer srv.Close()

		resp, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		_, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err == nil {
			t.Fatal("expected handshake read to fail due to timeout, got nil")
		}
	})

	t.Run("legacy SSE handshake that sends a partial event then stalls times out", func(t *testing.T) {
		client := timeoutClient(agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_SSE)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			f := w.(http.Flusher)
			// Incomplete event: no blank-line terminator, so Connect can't return.
			_, _ = w.Write([]byte("event: endpoint\n"))
			f.Flush()
			<-r.Context().Done()
		}))
		defer srv.Close()

		resp, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		_, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err == nil {
			t.Fatal("expected partial-event handshake to time out, got nil")
		}
	})

	t.Run("legacy SSE stream unbounded after first event", func(t *testing.T) {
		client := timeoutClient(agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_SSE)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			f := w.(http.Flusher)
			// First event (endpoint) arrives promptly.
			_, _ = w.Write([]byte("event: endpoint\ndata: /msg\n\n"))
			f.Flush()
			// A later event arrives well past the handshake timeout.
			time.Sleep(300 * time.Millisecond)
			_, _ = w.Write([]byte("data: late\n\n"))
			f.Flush()
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
		if !strings.Contains(string(body), "late") {
			t.Fatalf("expected late event to be read, got %q", body)
		}
	})

	t.Run("message POST body that stalls after headers times out", func(t *testing.T) {
		client := timeoutClient(agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP)
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
		client := timeoutClient(agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP)
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
