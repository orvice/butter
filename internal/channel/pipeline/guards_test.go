package pipeline

import (
	"context"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestHandle_AdmissionReject_DoesNothing(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	msg := baseMsg()
	msg.Admission = []AdmissionRule{{Value: "1", Allowlist: []string{"99", "100"}}}

	h.Handle(context.Background(), msg)

	if len(r.runCalls) != 0 {
		t.Errorf("expected no runner call, got %d", len(r.runCalls))
	}
	if len(tr.replies) != 0 {
		t.Errorf("expected no reply, got %v", tr.replies)
	}
}

func TestHandle_AdmissionPass_Runs(t *testing.T) {
	h, r, _, _, _, _ := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	msg := baseMsg()
	msg.Admission = []AdmissionRule{
		{Value: "1", Allowlist: []string{"1", "2"}},
		{Value: "42", Allowlist: nil}, // empty allowlist = no restriction
	}

	h.Handle(context.Background(), msg)

	if len(r.runCalls) != 1 {
		t.Errorf("expected 1 runner call, got %d", len(r.runCalls))
	}
}

func TestHandle_AdmissionSkipWhenEmpty(t *testing.T) {
	h, r, _, _, _, _ := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	msg := baseMsg()
	// Value empty + SkipWhenEmpty: rule is skipped even though the allowlist
	// would otherwise reject an empty value (Discord guild allowlist in a DM).
	msg.Admission = []AdmissionRule{{Value: "", Allowlist: []string{"guild1"}, SkipWhenEmpty: true}}

	h.Handle(context.Background(), msg)

	if len(r.runCalls) != 1 {
		t.Errorf("expected 1 runner call, got %d", len(r.runCalls))
	}
}

func TestHandle_TriggerNoMatch_DoesNothing(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{
		ChannelName:  "tg",
		DefaultAgent: "assistant",
		Triggers: []*agentsv1.AgentTrigger{
			{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT},
		},
	})
	msg := baseMsg()
	msg.IsPrivate = false // group chat, private-chat trigger won't match

	h.Handle(context.Background(), msg)

	if len(r.runCalls) != 0 || len(tr.replies) != 0 {
		t.Errorf("expected nothing, got runs=%d replies=%v", len(r.runCalls), tr.replies)
	}
}

func TestHandle_TriggerMessageMatches(t *testing.T) {
	h, r, _, _, _, _ := newHarness(Config{
		ChannelName:  "tg",
		DefaultAgent: "assistant",
		Triggers: []*agentsv1.AgentTrigger{
			{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MESSAGE},
		},
	})

	h.Handle(context.Background(), baseMsg())

	if len(r.runCalls) != 1 {
		t.Errorf("expected 1 runner call, got %d", len(r.runCalls))
	}
}

func TestHandle_EmptyMessage_DoesNothing(t *testing.T) {
	h, r, _, _, _, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	msg := baseMsg()
	msg.Text = ""
	msg.HasMedia = false

	h.Handle(context.Background(), msg)

	if len(r.runCalls) != 0 || len(tr.replies) != 0 {
		t.Errorf("expected nothing, got runs=%d replies=%v", len(r.runCalls), tr.replies)
	}
}
