package main

import (
	"context"

	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	wsctx "go.orx.me/apps/butter/internal/workspace"
)

const (
	sessionsWorkspace = "sessions"
	sessionsPeerWS    = "sessions-peer"
	sessionsAgentID   = "helper"
	sessionsBobUser   = "bob"
)

// seedExplorerSessions creates ADK sessions used by session_contract_test and
// sessions_runtime_test.go. They share the chat fixture's in-memory store.
func seedExplorerSessions(ctx context.Context, store *chatSessions) error {
	if err := createSessionWithEvents(ctx, store, sessionsWorkspace, "web-chat", chatUserID, "sess-own", "hello own"); err != nil {
		return err
	}
	if err := createSessionWithEvents(ctx, store, sessionsWorkspace, "web-chat", sessionsBobUser, "sess-bob", "hello bob"); err != nil {
		return err
	}
	if err := createSessionWithEvents(ctx, store, sessionsWorkspace, "cron:nightly", "cron:nightly", "sess-cron", "cron question"); err != nil {
		return err
	}
	return createSessionWithEvents(ctx, store, sessionsPeerWS, "web-chat", chatUserID, "sess-peer", "peer ws")
}

func createSessionWithEvents(ctx context.Context, store *chatSessions, wsID, app, user, id, userText string) error {
	wctx := wsctx.WithID(ctx, wsID)
	resp, err := store.Create(wctx, &adksession.CreateRequest{AppName: app, UserID: user, SessionID: id})
	if err != nil {
		return err
	}
	userEv := adksession.NewEvent(ctx, "inv-"+id)
	userEv.Author = "user"
	userEv.Content = genai.NewContentFromText(userText, genai.RoleUser)
	if err := store.AppendEvent(wctx, resp.Session, userEv); err != nil {
		return err
	}
	agentEv := adksession.NewEvent(ctx, "inv-"+id)
	agentEv.Author = "assistant"
	agentEv.Content = genai.NewContentFromText("fixture reply", genai.RoleModel)
	return store.AppendEvent(wctx, resp.Session, agentEv)
}
