package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/service"
)

func TestHealthHandler_Ping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := service.NewHealthService(repo.NewHealthRepository(), &config.AppConfig{
		HTTP: config.HTTPConfig{Greeting: "ok"},
	})
	NewHealthHandler(svc).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Service string `json:"service"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if body.Service != "butter" || body.Message != "ok" {
		t.Fatalf("body = %+v, want service=butter message=ok", body)
	}
}
