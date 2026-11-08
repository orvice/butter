package plane

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type activityRoundTripFunc func(*http.Request) (*http.Response, error)

func (f activityRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAmbiguousActivityResponseStaysUnknownWithoutResend(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"missing ID", 201, `{}`},
		{"blank ID", 201, `{"id":" "}`},
		{"malformed response", 201, `{"id":`},
		{"request timeout", 408, ``},
		{"conflict without creation guarantee", 409, ``},
		{"server failure", 500, ``},
		{"gateway lost upstream response", 502, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := NewActivityHTTPClient("https://plane.invalid", &http.Client{
				Transport: activityRoundTripFunc(func(*http.Request) (*http.Response, error) {
					calls++
					return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
				}),
			}, func(context.Context, string, string, string) (string, error) { return "fixture", nil })
			store := NewMemoryDeliveryStore()
			p := NewDurablePublisher(ActivityReplyPublisher{Client: client}, store)
			e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "prompt"}
			if err := p.PublishReply(t.Context(), e, "reply"); err == nil {
				t.Fatal("ambiguous POST acknowledged as delivered")
			}
			if err := p.Recover(t.Context()); err != nil {
				t.Fatal(err)
			}
			_ = p.PublishReply(t.Context(), e, "reply")
			row, err := store.Get(t.Context(), DeliveryKey(e))
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || row.Status != DeliveryUnknown || row.Attempts != 1 {
				t.Fatalf("calls=%d status=%s attempts=%d; want one POST and UNKNOWN", calls, row.Status, row.Attempts)
			}
		})
	}
}

func TestActivityResponseWithIDIsDelivered(t *testing.T) {
	c := NewActivityHTTPClient("https://plane.invalid", &http.Client{Transport: activityRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(`{"id":"activity-created"}`))}, nil
	})}, func(context.Context, string, string, string) (string, error) { return "fixture", nil })
	result, err := c.CreateActivity(t.Context(), ActivityRequest{})
	if err != nil || result.ID != "activity-created" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestPermanentActivityFailureDoesNotAutomaticallyResend(t *testing.T) {
	store := NewMemoryDeliveryStore()
	target := &deliveryTarget{err: &activityError{kind: DeliveryPermanentFailure, err: errors.New("forbidden")}}
	p := NewDurablePublisher(target, store)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "prompt"}
	_ = p.PublishReply(t.Context(), e, "reply")
	_ = p.PublishReply(t.Context(), e, "reply")
	if target.calls != 1 {
		t.Fatalf("calls=%d, want no automatic retry of permanent failure", target.calls)
	}
}
