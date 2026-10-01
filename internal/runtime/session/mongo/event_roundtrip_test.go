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

func sampleEvent() *session.Event {
	return &session.Event{
		ID:             "event-1",
		Timestamp:      time.Date(2026, 8, 12, 10, 11, 12, 0, time.UTC),
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
}

func TestEventDocRoundTripPreservesADKEvent(t *testing.T) {
	evt := sampleEvent()

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

// adkV21EventJSON is sampleEvent as ADK v2.1.0 encoded it, verbatim. Until
// v2.2.0 gave session.Event and model.LLMResponse camelCase JSON tags, the
// keys were Go field names, and events stored before the upgrade keep them.
const adkV21EventJSON = `{
	"Content": {
		"parts": [
			{
				"text": "done"
			}
		],
		"role": "model"
	},
	"CitationMetadata": null,
	"GroundingMetadata": null,
	"UsageMetadata": {
		"candidatesTokenCount": 2,
		"promptTokenCount": 7
	},
	"CustomMetadata": {
		"response_id": "resp-1"
	},
	"LogprobsResult": null,
	"InputTranscription": null,
	"OutputTranscription": null,
	"ModelVersion": "model-v1",
	"Partial": false,
	"TurnComplete": true,
	"Interrupted": true,
	"SessionResumptionHandle": "",
	"ErrorCode": "",
	"ErrorMessage": "",
	"FinishReason": "MAX_TOKENS",
	"AvgLogprobs": -0.25,
	"ID": "event-1",
	"Timestamp": "2026-08-12T10:11:12Z",
	"InvocationID": "inv-1",
	"Branch": "root.worker",
	"isolationScope": "task-1",
	"Author": "worker",
	"Actions": {
		"StateDelta": {
			"phase": "done"
		},
		"ArtifactDelta": {
			"report.txt": 3
		},
		"RequestedToolConfirmations": null,
		"SkipSummarization": true,
		"TransferToAgent": "reviewer",
		"Escalate": true
	},
	"LongRunningToolIDs": [
		"call-1"
	],
	"Routes": [
		"approved"
	],
	"RequestedInput": {
		"interruptId": "ask-1",
		"message": "Approve?",
		"payload": {
			"document": "draft"
		}
	},
	"Output": {
		"status": "ok"
	},
	"nodeInfo": {
		"path": "review",
		"messageAsOutput": true,
		"outputFor": [
			"review"
		]
	}
}`

func TestEventFromDocReadsADKV21EventJSON(t *testing.T) {
	want := sampleEvent()
	doc := eventDoc{
		EventID:      want.ID,
		InvocationID: want.InvocationID,
		Author:       want.Author,
		Branch:       want.Branch,
		EventJSON:    []byte(adkV21EventJSON),
		Timestamp:    want.Timestamp,
	}

	got, err := eventFromDoc(context.Background(), doc)
	if err != nil {
		t.Fatalf("eventFromDoc: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("v2.1.0 event not restored:\n got: %#v\nwant: %#v", got, want)
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
