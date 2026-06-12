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
		if err == nil || !strings.Contains(err.Error(), "no response headers") {
			t.Fatalf("expected header timeout error, got %v", err)
		}
	})

	t.Run("streaming body after headers is unaffected", func(t *testing.T) {
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
}
