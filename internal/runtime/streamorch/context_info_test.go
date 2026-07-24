package streamorch

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewContextInfo_DefaultsAppNameAndUserID(t *testing.T) {
	ctxInfo, err := NewContextInfo(ContextInfoInput{
		WorkspaceID:  "ws-1",
		HasWorkspace: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctxInfo.GetChannelName() != "api" {
		t.Fatalf("expected default channel_name %q, got %q", "api", ctxInfo.GetChannelName())
	}
	if ctxInfo.GetUserId() != "api" {
		t.Fatalf("expected default user_id %q, got %q", "api", ctxInfo.GetUserId())
	}
}

func TestNewContextInfo_GeneratesSessionIDWithPrefixWhenEmpty(t *testing.T) {
	ctxInfo, err := NewContextInfo(ContextInfoInput{
		WorkspaceID:   "ws-1",
		HasWorkspace:  true,
		SessionPrefix: "chat-",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(ctxInfo.GetSessionId(), "chat-") {
		t.Fatalf("expected generated session_id to have prefix %q, got %q", "chat-", ctxInfo.GetSessionId())
	}
	if len(ctxInfo.GetSessionId()) == len("chat-") {
		t.Fatalf("expected a generated suffix after the prefix, got %q", ctxInfo.GetSessionId())
	}
}

func TestNewContextInfo_GeneratesInvocationUUID(t *testing.T) {
	ctxInfo, err := NewContextInfo(ContextInfoInput{
		WorkspaceID:  "ws-1",
		HasWorkspace: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := uuid.Parse(ctxInfo.GetUuid()); err != nil {
		t.Fatalf("expected uuid-shaped invocation id, got %q: %v", ctxInfo.GetUuid(), err)
	}
}

func TestNewContextInfo_ErrorsWhenNoWorkspaceAndNotAdmin(t *testing.T) {
	_, err := NewContextInfo(ContextInfoInput{
		HasWorkspace: false,
		IsAdmin:      false,
	})
	if err == nil {
		t.Fatal("expected an error when no workspace is set and caller is not admin")
	}
}

func TestNewContextInfo_AdminAllowedWithoutWorkspace(t *testing.T) {
	ctxInfo, err := NewContextInfo(ContextInfoInput{
		HasWorkspace: false,
		IsAdmin:      true,
	})
	if err != nil {
		t.Fatalf("unexpected error for admin caller: %v", err)
	}
	if ctxInfo.GetWorkspaceId() != "" {
		t.Fatalf("expected empty workspace id for admin system path, got %q", ctxInfo.GetWorkspaceId())
	}
}

func TestNewContextInfo_PreservesExplicitSessionID(t *testing.T) {
	ctxInfo, err := NewContextInfo(ContextInfoInput{
		WorkspaceID:   "ws-1",
		HasWorkspace:  true,
		SessionID:     "explicit-session",
		SessionPrefix: "chat-",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctxInfo.GetSessionId() != "explicit-session" {
		t.Fatalf("expected explicit session_id to be preserved, got %q", ctxInfo.GetSessionId())
	}
}
