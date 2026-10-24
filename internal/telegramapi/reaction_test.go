package telegramapi

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"testing"
)

func TestSetMessageReactionPayload(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot"+testToken+"/setMessageReaction" {
			t.Errorf("unexpected method: %s", r.URL.Path)
		}
		var payload struct {
			ChatID    string `json:"chat_id"`
			MessageID string `json:"message_id"`
			Reaction  []struct {
				Type  string `json:"type"`
				Emoji string `json:"emoji"`
			} `json:"reaction"`
		}
		if err := json.UnmarshalRead(r.Body, &payload); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if payload.ChatID != "-1001234567890" || payload.MessageID != "9" {
			t.Errorf("wrong reaction target: %+v", payload)
		}
		if len(payload.Reaction) != 1 || payload.Reaction[0].Type != "emoji" || payload.Reaction[0].Emoji != "👀" {
			t.Errorf("wrong reaction: %+v", payload.Reaction)
		}
		fmt.Fprint(w, `{"ok":true,"result":true}`)
	})
	if err := client.SetMessageReaction(t.Context(), ReactionParams{
		ChatID: "-1001234567890", MessageID: "9", Emoji: "👀",
	}); err != nil {
		t.Fatalf("SetMessageReaction: %v", err)
	}
}

func TestSetMessageReactionReportsRejection(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"ok":false,"error_code":400,"description":"Bad Request: REACTION_INVALID"}`)
	})
	if err := client.SetMessageReaction(t.Context(), ReactionParams{
		ChatID: "-1001234567890", MessageID: "9", Emoji: "👀",
	}); err == nil {
		t.Fatal("expected a reaction error")
	}
}
