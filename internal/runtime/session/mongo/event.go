package mongo

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

func eventToDoc(appName, sessionID string, evt *session.Event) (eventDoc, error) {
	eventJSON, err := json.Marshal(evt)
	if err != nil {
		return eventDoc{}, fmt.Errorf("marshaling event: %w", err)
	}

	doc := eventDoc{
		SessionID:    sessionID,
		AppName:      appName,
		EventID:      evt.ID,
		InvocationID: evt.InvocationID,
		Author:       evt.Author,
		Branch:       evt.Branch,
		EventJSON:    eventJSON,
		Timestamp:    evt.Timestamp,
	}
	// Keep writing the legacy field while older tooling still reads it.
	if evt.Content != nil {
		doc.ContentJSON, err = json.Marshal(evt.Content)
		if err != nil {
			return eventDoc{}, fmt.Errorf("marshaling event content: %w", err)
		}
	}
	return doc, nil
}

func eventFromDoc(ctx context.Context, doc eventDoc) (*session.Event, error) {
	if len(doc.EventJSON) > 0 {
		var evt session.Event
		if err := json.Unmarshal(doc.EventJSON, &evt); err == nil {
			applyEventEnvelope(&evt, doc)
			return &evt, nil
		}
	}

	evt := session.NewEvent(ctx, doc.InvocationID)
	applyEventEnvelope(evt, doc)
	if len(doc.ContentJSON) == 0 {
		if len(doc.EventJSON) > 0 {
			return evt, fmt.Errorf("unmarshaling event JSON")
		}
		return evt, nil
	}

	var content genai.Content
	if err := json.Unmarshal(doc.ContentJSON, &content); err != nil {
		return evt, fmt.Errorf("unmarshaling legacy event content: %w", err)
	}
	evt.LLMResponse = model.LLMResponse{Content: &content}
	if len(doc.EventJSON) > 0 {
		return evt, fmt.Errorf("unmarshaling event JSON; restored legacy content only")
	}
	return evt, nil
}

func applyEventEnvelope(evt *session.Event, doc eventDoc) {
	evt.ID = doc.EventID
	evt.Timestamp = doc.Timestamp
	evt.InvocationID = doc.InvocationID
	evt.Author = doc.Author
	evt.Branch = doc.Branch
}
