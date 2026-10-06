package application

import (
	"context"
	"testing"

	"google.golang.org/adk/v2/session"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// A turn hook calls TitleSession on a context that carries no caller: no
// signed-in user and no admin. Hooks used to go through GenerateSessionTitle,
// whose authorization answered Unauthenticated, so no title was ever stored.
// TitleSession must title an untitled session on that context and leave a
// titled one alone.
func TestTitleSession_TitlesOnABackgroundContext(t *testing.T) {
	cases := []struct {
		name string
		// sessionTitle is the session's title when the turn completes;
		// storeTitle one the store already holds by the time the CAS runs
		// (a rename racing the hook).
		sessionTitle string
		storeTitle   string

		wantModelCall bool
		wantCAS       int
		wantStored    string // "" when no title may be written
	}{
		{
			name:          "untitled session gets the model's title",
			wantModelCall: true, wantCAS: 1, wantStored: "Kyoto in April",
		},
		{
			name:         "titled session is left alone",
			sessionTitle: "Manual title",
		},
		{
			name:          "a title written meanwhile wins",
			storeTitle:    "Renamed meanwhile",
			wantModelCall: true, wantCAS: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := &fakeSessionWithEvents{
				fakeSession: fakeSession{id: "s1", title: tc.sessionTitle, state: &fakeState{data: map[string]any{}}},
				events: []*session.Event{
					makeEvent("user", textPart("Plan a trip to Kyoto in April")),
					makeEvent("agent", textPart("Here is a plan for your trip.")),
				},
			}
			store := &stubTitleStore{existingTitle: tc.storeTitle}
			llm := &fakeLLM{response: textResponse("Kyoto in April")}
			svc := newLLMTitleTestService(store,
				llmAgentResolver("ws1", "flash"),
				&fakeProviderLister{providers: map[string][]*agentsv1.ModelProvider{
					"ws1": {providerWithAlias("p1", "flash", "gemini-2.0-flash")},
				}},
				"", sess)
			svc.titleResolveModel = (&fakeResolve{llm: llm}).fn

			if _, _, err := svc.TitleSession(context.Background(), "web-chat", "u1", "s1"); err != nil {
				t.Fatalf("TitleSession: %v", err)
			}

			if llm.called != tc.wantModelCall {
				t.Errorf("title model called = %v, want %v", llm.called, tc.wantModelCall)
			}
			if store.casCalled != tc.wantCAS {
				t.Fatalf("CAS calls = %d, want %d", store.casCalled, tc.wantCAS)
			}
			if tc.wantCAS > 0 && store.casSession != "web-chat/u1/s1" {
				t.Errorf("CAS addressed %q, want the session web-chat/u1/s1", store.casSession)
			}
			if store.lastTitle != tc.wantStored {
				t.Errorf("stored title = %q, want %q", store.lastTitle, tc.wantStored)
			}
		})
	}
}
