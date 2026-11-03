package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWorkStationConnectRouteAndNoStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, _ := SetupRoutes(&config.AppConfig{Auth: config.AuthConfig{AllowUnauthenticated: true}}, daemon.NewRegistry())
	engine := gin.New()
	router(engine)
	req := httptest.NewRequest(http.MethodPost, "/api/agents.v1.WorkStationService/ListConnections", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workspace-ID", "A")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("route not wired: %d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private projection cacheable")
	}
}

func TestWorkStationShutdownCancelsAndJoins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	h := &Handlers{stopWorkStation: cancel, workStationDone: done}
	go func() { <-ctx.Done(); close(done) }()
	deadline, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := h.ShutdownWorkStation(deadline); err != nil {
		t.Fatal(err)
	}
	if err := h.ShutdownWorkStation(deadline); err != nil {
		t.Fatal("shutdown not idempotent", err)
	}
}
