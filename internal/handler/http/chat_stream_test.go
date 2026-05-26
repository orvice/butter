package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestChatStreamRequiresWorkspaceForNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewChatStreamHandler()
	h.SetRunnerService(&runner.Service{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := auth.WithAuthenticated(c.Request.Context(), &agentsv1.User{Id: "u1", Role: "member"}, &auth.Session{})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	h.Register(r)

	body := bytes.NewBufferString(`{"agent_name":"assistant","message":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/chat/stream", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestEventTextPartsReturnsPartialTextDeltas(t *testing.T) {
	evt := testEvent(true, []*genai.Part{
		{Text: "Hel"},
		{Text: "thinking", Thought: true},
		{Text: "lo"},
		{FunctionCall: &genai.FunctionCall{Name: "tool"}},
	})

	got := eventTextParts(evt)
	want := []string{"Hel", "lo"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("eventTextParts() = %#v, want %#v", got, want)
	}
}

func TestEventTextPartsIgnoresNonPartialEvents(t *testing.T) {
	evt := testEvent(false, []*genai.Part{{Text: "final"}})

	if got := eventTextParts(evt); got != nil {
		t.Fatalf("eventTextParts() = %#v, want nil", got)
	}
}

func TestEventHasOnlyTextParts(t *testing.T) {
	tests := []struct {
		name string
		evt  *session.Event
		want bool
	}{
		{
			name: "text only",
			evt: testEvent(true, []*genai.Part{
				{Text: "Hel"},
				{Text: "lo"},
			}),
			want: true,
		},
		{
			name: "text plus tool call",
			evt: testEvent(true, []*genai.Part{
				{Text: "Checking"},
				{FunctionCall: &genai.FunctionCall{Name: "lookup"}},
			}),
			want: false,
		},
		{
			name: "empty content",
			evt:  testEvent(true, nil),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := eventHasOnlyTextParts(tt.evt); got != tt.want {
				t.Fatalf("eventHasOnlyTextParts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testEvent(partial bool, parts []*genai.Part) *session.Event {
	return &session.Event{
		LLMResponse: model.LLMResponse{
			Partial: partial,
			Content: &genai.Content{Parts: parts, Role: genai.RoleModel},
		},
	}
}
