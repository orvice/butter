package mongo

import (
	"context"
	"reflect"
	"testing"
	"time"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

func TestEventDocRoundTripPreservesADKEvent(t *testing.T) {
	timestamp := time.Date(2026, 8, 12, 10, 11, 12, 0, time.UTC)
	evt := &session.Event{
		ID:             "event-1",
		Timestamp:      timestamp,
		InvocationID:   "inv-1",
		Branch:         "root.worker",
		IsolationScope: "task-1",
		Author:         "worker",
		Actions: session.EventActions{
			StateDelta:        map[string]any{"phase": "done"},
			ArtifactDelta:     map[string]int64{"report.txt": 3},
			SkipSummarization: true,
			TransferToAgent:   "reviewer",
			Escalate:          true,
		},
		LongRunningToolIDs: []string{"call-1"},
		Routes:             []string{"approved"},
		RequestedInput: &session.RequestInput{
			InterruptID: "ask-1",
			Message:     "Approve?",
			Payload:     map[string]any{"document": "draft"},
		},
		Output:   map[string]any{"status": "ok"},
		NodeInfo: &session.NodeInfo{Path: "review", MessageAsOutput: true, OutputFor: []string{"review"}},
		LLMResponse: model.LLMResponse{
			Content:        genai.NewContentFromText("done", genai.RoleModel),
			CustomMetadata: map[string]any{"response_id": "resp-1"},
			UsageMetadata:  &genai.GenerateContentResponseUsageMetadata{PromptTokenCount: 7, CandidatesTokenCount: 2},
			Partial:        false,
			TurnComplete:   true,
			ErrorCode:      "",
			ErrorMessage:   "",
			Interrupted:    true,
			FinishReason:   genai.FinishReasonMaxTokens,
			ModelVersion:   "model-v1",
			AvgLogprobs:    -0.25,
		},
	}

	doc, err := eventToDoc("app", "session-1", evt)
	if err != nil {
		t.Fatalf("eventToDoc: %v", err)
	}
	got, err := eventFromDoc(context.Background(), doc)
	if err != nil {
		t.Fatalf("eventFromDoc: %v", err)
	}

	if !reflect.DeepEqual(got, evt) {
		t.Fatalf("round trip mismatch:\n got: %#v\nwant: %#v", got, evt)
	}
}

func TestEventFromDocReadsLegacyContentJSON(t *testing.T) {
	doc := eventDoc{
		EventID:      "legacy-event",
		InvocationID: "legacy-inv",
		Author:       "assistant",
		Branch:       "root",
		ContentJSON:  []byte(`{"role":"model","parts":[{"text":"legacy reply"}]}`),
		Timestamp:    time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC),
	}

	got, err := eventFromDoc(context.Background(), doc)
	if err != nil {
		t.Fatalf("eventFromDoc: %v", err)
	}
	if got.Content == nil || len(got.Content.Parts) != 1 || got.Content.Parts[0].Text != "legacy reply" {
		t.Fatalf("legacy content not restored: %#v", got.Content)
	}
	if got.ID != doc.EventID || got.InvocationID != doc.InvocationID || got.Author != doc.Author || got.Branch != doc.Branch {
		t.Fatalf("legacy envelope not restored: %#v", got)
	}
}
