package interrupt

import (
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"
)

// TestResume_RewrapsTextToOldest: a plain-text reply with an interrupt pending
// is rewrapped as the FunctionResponse answering the oldest one.
func TestResume_RewrapsTextToOldest(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{
		askEvent("i1", "Question A?"),
		askEvent("i2", "Question B?"),
	}}
	resumed, ok := Resume(sess, []*genai.Part{{Text: "first answer"}})
	if !ok {
		t.Fatal("Resume did not rewrap a text answer with a pending Interrupt")
	}
	if len(resumed) != 1 {
		t.Fatalf("len(resumed) = %d, want 1", len(resumed))
	}
	fr := resumed[0].FunctionResponse
	if fr == nil {
		t.Fatal("first part is not a FunctionResponse")
	}
	if fr.ID != "i1" {
		t.Errorf("answered interrupt = %q, want i1 (oldest)", fr.ID)
	}
	if fr.Name != workflow.WorkflowInputFunctionCallName {
		t.Errorf("FunctionResponse name = %q, want %q", fr.Name, workflow.WorkflowInputFunctionCallName)
	}
	if got := fr.Response["payload"]; got != "first answer" {
		t.Errorf("payload = %v, want %q", got, "first answer")
	}
}

// TestResume_PassthroughWhenFunctionResponsePresent: a precisely-addressed
// reply (already a FunctionResponse) is left untouched.
func TestResume_PassthroughWhenFunctionResponsePresent(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{askEvent("i1", "Q?")}}
	parts := []*genai.Part{{FunctionResponse: &genai.FunctionResponse{ID: "i1"}}}
	resumed, ok := Resume(sess, parts)
	if ok {
		t.Error("Resume rewrapped a message that already carried a FunctionResponse")
	}
	if len(resumed) != 1 || resumed[0].FunctionResponse == nil {
		t.Errorf("parts changed: %+v", resumed)
	}
}

// TestResume_PassthroughWhenNothingPending: with no pending Interrupt the reply
// passes through as plain text.
func TestResume_PassthroughWhenNothingPending(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{
		askEvent("i1", "Q?"),
		answerEvent("i1"),
	}}
	parts := []*genai.Part{{Text: "hello"}}
	resumed, ok := Resume(sess, parts)
	if ok {
		t.Error("Resume rewrapped a reply with nothing pending")
	}
	if len(resumed) != 1 || resumed[0].Text != "hello" {
		t.Errorf("parts changed: %+v", resumed)
	}
}

// TestResume_PassthroughWhenNoText: a reply carrying no text (e.g. only an
// image) cannot answer a text interrupt and passes through.
func TestResume_PassthroughWhenNoText(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{askEvent("i1", "Q?")}}
	image := &genai.Part{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte{1, 2, 3}}}
	resumed, ok := Resume(sess, []*genai.Part{image})
	if ok {
		t.Error("Resume rewrapped a reply with no text")
	}
	if len(resumed) != 1 || resumed[0].InlineData == nil {
		t.Errorf("parts changed: %+v", resumed)
	}
}

// TestResume_PreservesNonTextParts: an answer sent with an attachment keeps the
// attachment — only the text becomes the answer; other parts ride along.
func TestResume_PreservesNonTextParts(t *testing.T) {
	sess := &fakeSession{events: []*session.Event{askEvent("i1", "Q?")}}
	image := &genai.Part{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte{1, 2, 3}}}
	resumed, ok := Resume(sess, []*genai.Part{{Text: "approved"}, image})
	if !ok {
		t.Fatal("Resume did not rewrap the text answer")
	}
	if resumed[0].FunctionResponse == nil {
		t.Fatal("first part is not the Interrupt's FunctionResponse")
	}
	found := false
	for _, p := range resumed {
		if p.InlineData != nil {
			found = true
		}
	}
	if !found {
		t.Error("the image part was dropped from the resumed message")
	}
}
