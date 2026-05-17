package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/service"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type statusStore struct {
	*configmemory.Store
}

func (s *statusStore) ActiveBackendName() string {
	return "memory"
}

func TestStatusHandler_Status(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := t.Context()
	cfg := &config.AppConfig{APIToken: "secret-token"}
	store := &statusStore{Store: configmemory.New()}
	store.Seed(ctx,
		[]agentsv1.Agent{{Name: "assistant"}},
		[]agentsv1.MCPServer{{Id: "mcp-primary"}},
		[]agentsv1.RemoteAgent{{Id: "remote-primary"}},
		[]agentsv1.AgentChannel{{Name: "telegram-main"}},
		nil,
	)

	r := gin.New()
	r.Use(APITokenAuthMiddleware(cfg, nil))
	NewStatusHandler(service.NewStatusService(cfg, store)).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/status", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	var body struct {
		Service string `json:"service"`
		Storage struct {
			ConfiguredBackend string `json:"configured_backend"`
			ActiveBackend     string `json:"active_backend"`
			Persistent        bool   `json:"persistent"`
			Collections       struct {
				Agents       int `json:"agents"`
				MCPServers   int `json:"mcp_servers"`
				RemoteAgents int `json:"remote_agents"`
				Channels     int `json:"channels"`
			} `json:"collections"`
		} `json:"storage"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if body.Service != "butter" {
		t.Fatalf("service = %q, want butter", body.Service)
	}
	if body.Storage.ConfiguredBackend != "memory" || body.Storage.ActiveBackend != "memory" || body.Storage.Persistent {
		t.Fatalf("storage = %+v, want memory active/configured and non-persistent", body.Storage)
	}
	if body.Storage.Collections.Agents != 1 ||
		body.Storage.Collections.MCPServers != 1 ||
		body.Storage.Collections.RemoteAgents != 1 ||
		body.Storage.Collections.Channels != 1 {
		t.Fatalf("collections = %+v, want all counts 1", body.Storage.Collections)
	}
}

func TestStatusHandler_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.AppConfig{APIToken: "secret-token"}
	store := &statusStore{Store: configmemory.New()}

	r := gin.New()
	r.Use(APITokenAuthMiddleware(cfg, nil))
	NewStatusHandler(service.NewStatusService(cfg, store)).Register(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
