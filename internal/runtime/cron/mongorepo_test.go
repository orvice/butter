package cron

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCronJobDocumentRoundTripPreservesAgentID(t *testing.T) {
	want := &agentsv1.CronJob{
		Name:        "daily-ticket",
		WorkspaceId: "workspace-1",
		Schedule:    "0 9 * * *",
		AgentName:   "TicketManager",
		AgentId:     "ticketmanager",
		Enabled:     true,
	}

	encoded, err := bson.Marshal(jobDocFromProto(want))
	if err != nil {
		t.Fatalf("marshal cron job document: %v", err)
	}
	var decoded cronJobDoc
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal cron job document: %v", err)
	}

	got := jobDocToProto(&decoded)
	if got.GetAgentId() != want.GetAgentId() {
		t.Fatalf("agent_id = %q, want %q", got.GetAgentId(), want.GetAgentId())
	}
}

func TestCronExecutionDocumentRoundTripPreservesAgentID(t *testing.T) {
	want := &agentsv1.CronExecution{
		Id:          "execution-1",
		WorkspaceId: "workspace-1",
		JobName:     "daily-ticket",
		AgentName:   "TicketManager",
		AgentId:     "ticketmanager",
		StartedAt:   timestamppb.Now(),
	}

	encoded, err := bson.Marshal(docFromProto(want))
	if err != nil {
		t.Fatalf("marshal cron execution document: %v", err)
	}
	var decoded executionDoc
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal cron execution document: %v", err)
	}

	got := docToProto(&decoded)
	if got.GetAgentId() != want.GetAgentId() {
		t.Fatalf("agent_id = %q, want %q", got.GetAgentId(), want.GetAgentId())
	}
}
