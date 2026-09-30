package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	linearprocessingmemory "go.orx.me/apps/butter/internal/repo/linearprocessing/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	linearruntime "go.orx.me/apps/butter/internal/runtime/linear"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeResender struct {
	calls int
	err   error
}

func (r *fakeResender) Resend(_ context.Context, workspaceID, id string) (*agentsv1.LinearProcessingRecord, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	return &agentsv1.LinearProcessingRecord{Id: id, WorkspaceId: workspaceID,
		Status: agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED}, nil
}

func newLinearProcessingFixture(t *testing.T) (*LinearProcessingServiceServer, *linearprocessingmemory.Store, *fakeResender) {
	t.Helper()
	wsRepo := workspacememory.New()
	if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: "ws-a", Name: "ws-a", Slug: "ws-a"}); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	for _, m := range []struct{ user, role string }{{"owner", "owner"}, {"member", "member"}} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{WorkspaceId: "ws-a", UserId: m.user, Role: m.role}); err != nil {
			t.Fatalf("seed member: %v", err)
		}
	}
	records := linearprocessingmemory.New()
	now := time.Now()
	for _, d := range []struct {
		app, delivery string
		status        agentsv1.LinearProcessingStatus
	}{
		{"app-1", "d-1", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED},
		{"app-1", "d-2", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED},
		{"app-2", "d-3", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED},
	} {
		record, _, err := records.Claim(t.Context(), &agentsv1.LinearProcessingRecord{
			WorkspaceId: "ws-a", AppId: d.app, DeliveryId: d.delivery,
		}, "", now, now)
		if err != nil {
			t.Fatalf("seed record: %v", err)
		}
		record.Status = d.status
		if _, err := records.Update(t.Context(), record); err != nil {
			t.Fatalf("seed update: %v", err)
		}
	}
	resender := &fakeResender{}
	svc := NewLinearProcessingServiceServer(records)
	svc.SetWorkspaceRepo(wsRepo)
	svc.SetResender(resender)
	return svc, records, resender
}

func TestMembersListAndReadLinearProcessingRecords(t *testing.T) {
	svc, _, _ := newLinearProcessingFixture(t)
	list := func(req *agentsv1.ListLinearProcessingRecordsRequest) []*agentsv1.LinearProcessingRecord {
		resp, err := svc.ListLinearProcessingRecords(memberA(), connect.NewRequest(req))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		return resp.Msg.GetRecords()
	}
	if got := list(&agentsv1.ListLinearProcessingRecordsRequest{}); len(got) != 3 {
		t.Fatalf("all = %d, want 3", len(got))
	}
	if got := list(&agentsv1.ListLinearProcessingRecordsRequest{AppId: "app-1"}); len(got) != 2 {
		t.Fatalf("by app = %d, want 2", len(got))
	}
	failed := list(&agentsv1.ListLinearProcessingRecordsRequest{Status: agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED})
	if len(failed) != 1 {
		t.Fatalf("failed = %d, want 1", len(failed))
	}
	got, err := svc.GetLinearProcessingRecord(memberA(), connect.NewRequest(&agentsv1.GetLinearProcessingRecordRequest{Id: failed[0].GetId()}))
	if err != nil || got.Msg.GetRecord().GetDeliveryId() != "d-2" {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	if _, err := svc.GetLinearProcessingRecord(ctxAs("owner-b", "user", "ws-b"), connect.NewRequest(&agentsv1.GetLinearProcessingRecordRequest{Id: failed[0].GetId()})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("cross-workspace Get = %v, want NotFound", err)
	}
}

func TestOnlyOwnersAndAdminsResendLinearReplies(t *testing.T) {
	svc, _, resender := newLinearProcessingFixture(t)
	req := connect.NewRequest(&agentsv1.ResendLinearReplyRequest{Id: "any"})
	if _, err := svc.ResendLinearReply(memberA(), req); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("member resend = %v, want PermissionDenied", err)
	}
	if resender.calls != 0 {
		t.Fatal("a member's resend reached the resender")
	}
	if _, err := svc.ResendLinearReply(ownerA(), req); err != nil {
		t.Fatalf("owner resend: %v", err)
	}
	resender.err = linearruntime.ErrNotResendable
	if _, err := svc.ResendLinearReply(ownerA(), req); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("resend of a record without a reply = %v, want FailedPrecondition", err)
	}
	resender.err = errors.New("unreachable")
	if _, err := svc.ResendLinearReply(ownerA(), req); err == nil {
		t.Fatal("a failed resend reported success")
	}
}
