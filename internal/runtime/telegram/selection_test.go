package telegram

// Selection tests (issue #264/#269): who may switch, what a switch does to
// history, what a stale choice does, and how a reset behaves.

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// selectable builds a Destination where both Agent and Model selection are
// enabled and user 7 is a controller.
func selectable(mutate func(*agentsv1.TelegramDestinationConfig)) *agentsv1.TelegramDestination {
	return destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.AgentId = "support"
		c.Model = "fast"
		c.SelectableAgentIds = []string{"support", "research"}
		c.SelectableModels = []string{"fast", "pro"}
		c.ControllerUserIds = []string{"7"}
		if mutate != nil {
			mutate(c)
		}
	})
}

// command builds a bot_command message update.
func command(from, text string) string {
	slash := strings.SplitN(text, " ", 2)[0]
	return message(from, text,
		fmt.Sprintf(`"entities":[{"type":"bot_command","offset":0,"length":%d}]`, len([]rune(slash))))
}

func TestStoredAgentSelectionMovesTheSession(t *testing.T) {
	dest := selectable(nil)

	base := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername, Preferences{})
	switched := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername,
		Preferences{AgentID: "research"})

	if switched.AgentID != "research" {
		t.Fatalf("agent = %q", switched.AgentID)
	}
	if base.SessionID == switched.SessionID {
		t.Fatal("switching agent must move the session so histories stay separate")
	}
	// Switching back resumes the earlier conversation.
	back := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername,
		Preferences{AgentID: "support"})
	if back.SessionID != base.SessionID {
		t.Fatalf("switching back gave session %q, want %q", back.SessionID, base.SessionID)
	}
}

// A model change must not silently start a new conversation.
func TestStoredModelSelectionKeepsTheSession(t *testing.T) {
	dest := selectable(nil)

	base := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername, Preferences{})
	switched := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername,
		Preferences{Model: "pro"})

	if switched.Model != "pro" {
		t.Fatalf("model = %q", switched.Model)
	}
	if switched.SessionID != base.SessionID {
		t.Fatalf("switching model moved the session from %q to %q", base.SessionID, switched.SessionID)
	}
}

// A stored choice current configuration no longer allows falls back to the
// default immediately, and is reported so the caller can clear it.
func TestStaleSelectionsFallBackToTheDefault(t *testing.T) {
	dest := selectable(func(c *agentsv1.TelegramDestinationConfig) {
		c.SelectableAgentIds = []string{"support"}
		c.SelectableModels = []string{"fast"}
	})

	decision := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername,
		Preferences{AgentID: "research", Model: "pro"})

	if decision.AgentID != "support" || decision.Model != "fast" {
		t.Fatalf("effective routing = %s/%s, want the destination defaults",
			decision.AgentID, decision.Model)
	}
	if !decision.StaleSelection {
		t.Error("expected the stale selection to be reported so it can be cleared")
	}
}

// An empty selectable list locks selection, so nothing can override the
// default — that is what keeps topic routing fixed by default.
func TestEmptySelectableListsLockRouting(t *testing.T) {
	dest := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.AgentId = "support"
		c.Model = "fast"
	})

	decision := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername,
		Preferences{AgentID: "research", Model: "pro"})

	if decision.AgentID != "support" || decision.Model != "fast" {
		t.Fatalf("a locked destination accepted a selection: %s/%s", decision.AgentID, decision.Model)
	}
	if AgentSelectionEnabled(dest.GetConfig()) || ModelSelectionEnabled(dest.GetConfig()) {
		t.Error("selection must be reported as disabled")
	}
}

// --- Command behavior ------------------------------------------------------

func TestAgentCommandRequiresAController(t *testing.T) {
	fx := newSelectionFixture(t, nil)

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(command(`{"id":11,"is_bot":false}`, "/agent research"))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	stored, _ := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId()))
	if stored.AgentID != "" {
		t.Fatalf("a non-controller switched the agent to %q", stored.AgentID)
	}
}

func TestControllerSwitchesAndResetsTheAgent(t *testing.T) {
	fx := newSelectionFixture(t, nil)
	key := PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId())

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(command(realUser, "/agent research"))); err != nil {
		t.Fatalf("switch: %v", err)
	}
	stored, _ := fx.prefs.Get(t.Context(), key)
	if stored.AgentID != "research" {
		t.Fatalf("stored agent = %q", stored.AgentID)
	}

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(command(realUser, "/agent reset"))); err != nil {
		t.Fatalf("reset: %v", err)
	}
	stored, _ = fx.prefs.Get(t.Context(), key)
	if stored.AgentID != "" {
		t.Fatalf("reset left agent %q", stored.AgentID)
	}
}

func TestPiAgentLocksModelSelectionAndResumesItsSession(t *testing.T) {
	fx := newSelectionFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.SelectableAgentIds = []string{"support", "research", "pi-coder"}
	})
	fx.agents.known["pi-coder"] = "Pi Coder"
	fx.agents.modelOverrideLocked = map[string]bool{"pi-coder": true}

	handle := func(raw string) {
		t.Helper()
		if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
			t.Fatalf("Handle: %v", err)
		}
	}

	handle(command(realUser, "/agent pi-coder"))
	handle(command(realUser, "/model pro"))
	stored, err := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId()))
	if err != nil {
		t.Fatalf("read preferences: %v", err)
	}
	if stored.Model != "" {
		t.Fatalf("Pi agent stored a Butter model override %q", stored.Model)
	}
	lastReply := fx.bots.Sent()[len(fx.bots.Sent())-1]
	if !strings.Contains(lastReply.Params.Text, "Pi agents") {
		t.Fatalf("model lock reply = %q, want Pi-specific guidance", lastReply.Params.Text)
	}

	handle(message(realUser, "first pi turn", ""))
	handle(command(realUser, "/agent research"))
	handle(message(realUser, "research turn", ""))
	handle(command(realUser, "/agent pi-coder"))
	handle(message(realUser, "second pi turn", ""))

	if len(fx.agents.calls) != 3 {
		t.Fatalf("agent calls = %d, want 3", len(fx.agents.calls))
	}
	firstPi, research, secondPi := fx.agents.calls[0], fx.agents.calls[1], fx.agents.calls[2]
	if firstPi.agentName != "Pi Coder" || firstPi.model != "" {
		t.Fatalf("first Pi call = %+v, want no model override", firstPi)
	}
	if research.agentName != "Research Agent" || research.model != "fast" {
		t.Fatalf("research call = %+v, want destination model", research)
	}
	if secondPi.agentName != "Pi Coder" || secondPi.model != "" {
		t.Fatalf("second Pi call = %+v, want no model override", secondPi)
	}
	if secondPi.sessionID != firstPi.sessionID {
		t.Fatalf("Pi session changed after switching away and back: %q != %q", secondPi.sessionID, firstPi.sessionID)
	}
}

func TestSwitchingToANonCandidateIsRefused(t *testing.T) {
	fx := newSelectionFixture(t, nil)

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(command(realUser, "/agent nonexistent"))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	stored, _ := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId()))
	if stored.AgentID != "" {
		t.Fatalf("a non-candidate was accepted: %q", stored.AgentID)
	}
	if len(fx.bots.Sent()) != 1 {
		t.Fatal("expected the refusal to be explained")
	}
}

// A button rendered before a permission change must not still work.
func TestStaleCallbackIsRevalidated(t *testing.T) {
	fx := newSelectionFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		// "research" is no longer selectable, but an old button still offers it.
		c.SelectableAgentIds = []string{"support"}
	})

	raw := fmt.Sprintf(`{"update_id":1,"callback_query":{"id":"cb-1","from":%s,"data":"agent:research","message":{"message_id":9,"message_thread_id":42,"is_topic_message":true,"chat":{"id":-100,"type":"supergroup"},"from":{"id":111111,"is_bot":true}}}}`, realUser)
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	stored, _ := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId()))
	if stored.AgentID != "" {
		t.Fatalf("a stale button switched the agent to %q", stored.AgentID)
	}
}

func TestCallbackFromANonControllerIsRefused(t *testing.T) {
	fx := newSelectionFixture(t, nil)

	raw := `{"update_id":1,"callback_query":{"id":"cb-1","from":{"id":11,"is_bot":false},"data":"agent:research","message":{"message_id":9,"message_thread_id":42,"is_topic_message":true,"chat":{"id":-100,"type":"supergroup"},"from":{"id":111111,"is_bot":true}}}}`
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	stored, _ := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId()))
	if stored.AgentID != "" {
		t.Fatal("a non-controller callback changed shared routing")
	}
}

// A valid button press does take effect.
func TestValidCallbackAppliesTheSelection(t *testing.T) {
	fx := newSelectionFixture(t, nil)

	raw := fmt.Sprintf(`{"update_id":1,"callback_query":{"id":"cb-1","from":%s,"data":"agent:research","message":{"message_id":9,"message_thread_id":42,"is_topic_message":true,"chat":{"id":-100,"type":"supergroup"},"from":{"id":111111,"is_bot":true}}}}`, realUser)
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(raw)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	stored, _ := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId()))
	if stored.AgentID != "research" {
		t.Fatalf("stored agent = %q", stored.AgentID)
	}
}

// A stored choice that is no longer allowed is cleared, not just ignored.
func TestStaleSelectionIsClearedOnUse(t *testing.T) {
	fx := newSelectionFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.SelectableAgentIds = []string{"support"}
	})
	key := PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId())
	if err := fx.prefs.Put(t.Context(), key, Preferences{AgentID: "research"}); err != nil {
		t.Fatalf("seed preference: %v", err)
	}

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(message(realUser, "hello", ""))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	stored, _ := fx.prefs.Get(t.Context(), key)
	if stored.AgentID != "" {
		t.Fatalf("stale selection survived: %q", stored.AgentID)
	}
	if fx.agents.calls[0].agentName != "Support Agent" {
		t.Errorf("ran %q, want the destination default", fx.agents.calls[0].agentName)
	}
}

// Each user keeps their own selection under USER policy.
func TestPerUserSelectionIsIsolated(t *testing.T) {
	fx := newSelectionFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.SessionPolicy = agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_USER
		c.ControllerUserIds = []string{"7", "11"}
	})

	if err := fx.orchestrator.Handle(t.Context(),
		fx.eventForStored(command(realUser, "/agent research"))); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	mine, _ := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "u7"))
	theirs, _ := fx.prefs.Get(t.Context(), PreferenceKey(fx.dest.GetId(), "u11"))
	if mine.AgentID != "research" {
		t.Fatalf("my selection = %q", mine.AgentID)
	}
	if theirs.AgentID != "" {
		t.Fatalf("another user inherited the selection: %q", theirs.AgentID)
	}
}

// Preference lookup follows the accepted session policy, so editing a
// Destination while an event waits in Redis cannot make that event inherit a
// different user's selection.
func TestQueuedInteractionLoadsPreferencesForTheAcceptedSessionPolicy(t *testing.T) {
	fx := newSelectionFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.SessionPolicy = agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_DESTINATION
	})
	event := fx.eventForStored(message(realUser, "hello", ""))

	current := proto.Clone(fx.dest).(*agentsv1.TelegramDestination)
	current.Config.SessionPolicy = agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_USER
	updated, err := fx.repo.UpdateDestination(t.Context(), "ws-a", current, fx.dest.GetRevision())
	if err != nil {
		t.Fatalf("change current session policy: %v", err)
	}
	fx.dest = updated

	acceptedKey := PreferenceKey(fx.dest.GetId(), "d"+fx.dest.GetId())
	currentKey := PreferenceKey(fx.dest.GetId(), "u7")
	if err := fx.prefs.Put(t.Context(), acceptedKey, Preferences{AgentID: "research"}); err != nil {
		t.Fatalf("seed accepted-policy preference: %v", err)
	}
	if err := fx.prefs.Put(t.Context(), currentKey, Preferences{AgentID: "support"}); err != nil {
		t.Fatalf("seed current-policy preference: %v", err)
	}

	if err := fx.orchestrator.Handle(t.Context(), event); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(fx.agents.calls) != 1 || fx.agents.calls[0].agentName != "Research Agent" {
		t.Fatalf("agent calls = %+v, want the accepted-policy selection", fx.agents.calls)
	}
}

// newSelectionFixture reuses the orchestrator fixture with selection enabled.
func newSelectionFixture(t *testing.T, mutate func(*agentsv1.TelegramDestinationConfig)) *orchestratorFixture {
	t.Helper()
	fx := newOrchestratorFixture(t, func(c *agentsv1.TelegramDestinationConfig) {
		c.AgentId = "support"
		c.Model = "fast"
		c.SelectableAgentIds = []string{"support", "research"}
		c.SelectableModels = []string{"fast", "pro"}
		c.ControllerUserIds = []string{"7"}
		if mutate != nil {
			mutate(c)
		}
	})
	fx.agents.known["research"] = "Research Agent"
	return fx
}
