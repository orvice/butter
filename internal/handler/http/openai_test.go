package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// stubAgentRepo is a minimal in-memory AgentRepository for tests.
type stubAgentRepo struct {
	configrepo.AgentRepository
	agents []*agentsv1.Agent
}

func (s *stubAgentRepo) ListAgents(_ context.Context, _ string) ([]*agentsv1.Agent, error) {
	return s.agents, nil
}

func setupOpenAIRouter(repo configrepo.AgentRepository, authenticated bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	if authenticated {
		// Inject workspace context as if auth middleware ran.
		r.Use(func(c *gin.Context) {
			ctx := wsctx.WithID(c.Request.Context(), "test-workspace")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
	} else {
		// Use real auth middleware that requires a token.
		cfg := &config.AppConfig{APIToken: "secret-token"}
		r.Use(APITokenAuthMiddleware(cfg, nil))
	}

	h := NewOpenAIHandler(repo)
	h.Register(r)
	return r
}

func TestOpenAIModels_FiltersEnabledAgents(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "agent-enabled", EnableOpenaiApi: true},
			{Name: "agent-disabled", EnableOpenaiApi: false},
			{Name: "agent-unset"},
		},
	}
	router := setupOpenAIRouter(repo, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/models", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected object 'list', got %q", resp.Object)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 model, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "agent-enabled" {
		t.Errorf("expected model id 'agent-enabled', got %q", resp.Data[0].ID)
	}
	if resp.Data[0].Object != "model" {
		t.Errorf("expected object 'model', got %q", resp.Data[0].Object)
	}
	if resp.Data[0].OwnedBy != "butter" {
		t.Errorf("expected owned_by 'butter', got %q", resp.Data[0].OwnedBy)
	}
}

func TestOpenAIModels_EmptyWhenNoneEnabled(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "agent-a", EnableOpenaiApi: false},
			{Name: "agent-b"},
		},
	}
	router := setupOpenAIRouter(repo, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/models", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Object string        `json:"object"`
		Data   []interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("expected object 'list', got %q", resp.Object)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected empty data array, got %d items", len(resp.Data))
	}
}

func TestOpenAIModels_Unauthenticated(t *testing.T) {
	repo := &stubAgentRepo{
		agents: []*agentsv1.Agent{
			{Name: "agent-enabled", EnableOpenaiApi: true},
		},
	}
	router := setupOpenAIRouter(repo, false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/models", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
