package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"

	butterboxmemory "go.orx.me/apps/butter/internal/repo/butterbox/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakePiService is a typed fake of the butter-box PiService: only the RPCs
// the ButterBox service consumes are implemented, everything else answers
// Unimplemented. It records the bearer token of the last request.
type fakePiService struct {
	piv1connect.UnimplementedPiServiceHandler

	mu        sync.Mutex
	lastAuth  string
	sessions  []*piv1.Session
	models    []*piv1.Model
	lastModel *piv1.GetAvailableModelsRequest
}

func (f *fakePiService) recordAuth(header http.Header) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastAuth = header.Get("Authorization")
}

func (f *fakePiService) ListSessions(_ context.Context, req *connect.Request[piv1.ListSessionsRequest]) (*connect.Response[piv1.ListSessionsResponse], error) {
	f.recordAuth(req.Header())
	return connect.NewResponse(&piv1.ListSessionsResponse{Sessions: f.sessions}), nil
}

func (f *fakePiService) GetAvailableModels(_ context.Context, req *connect.Request[piv1.GetAvailableModelsRequest]) (*connect.Response[piv1.GetAvailableModelsResponse], error) {
	f.recordAuth(req.Header())
	f.mu.Lock()
	f.lastModel = req.Msg
	f.mu.Unlock()
	return connect.NewResponse(&piv1.GetAvailableModelsResponse{Models: f.models}), nil
}

type butterBoxFixture struct {
	svc  *ButterBoxServiceServer
	fake *fakePiService
	box  *httptest.Server
	ctx  context.Context
}

func newButterBoxFixture(t *testing.T) *butterBoxFixture {
	t.Helper()
	fake := &fakePiService{
		sessions: []*piv1.Session{{Id: "s1"}, {Id: "s2"}},
		models: []*piv1.Model{
			{Id: "claude-fable-5", Provider: "anthropic", Name: "Claude Fable 5", Reasoning: true},
			{Id: "gpt-5.6", Provider: "openai", Name: "GPT-5.6"},
		},
	}
	path, handler := piv1connect.NewPiServiceHandler(fake)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := NewButterBoxServiceServer(butterboxmemory.New())
	svc.SetKeyring(secretbox.NewKeyring(cryptokeymemory.New()))

	return &butterBoxFixture{
		svc:  svc,
		fake: fake,
		box:  srv,
		ctx:  workspace.WithID(t.Context(), "ws1"),
	}
}

func (f *butterBoxFixture) create(t *testing.T, name, token string) *agentsv1.ButterBox {
	t.Helper()
	resp, err := f.svc.CreateButterBox(f.ctx, connect.NewRequest(&agentsv1.CreateButterBoxRequest{
		Name:    name,
		BaseUrl: f.box.URL,
		Enabled: true,
		Token:   token,
	}))
	if err != nil {
		t.Fatalf("CreateButterBox: %v", err)
	}
	return resp.Msg.GetButterBox()
}

func TestButterBoxCRUD(t *testing.T) {
	f := newButterBoxFixture(t)

	created := f.create(t, "dev-box", "secret-token")
	if created.GetId() == "" {
		t.Fatal("no id assigned")
	}
	if !created.GetCredentialSet() {
		t.Fatal("credential_set false after create with token")
	}
	if created.GetWorkspaceId() != "ws1" {
		t.Fatalf("workspace_id = %q", created.GetWorkspaceId())
	}

	// Duplicate name is rejected.
	_, err := f.svc.CreateButterBox(f.ctx, connect.NewRequest(&agentsv1.CreateButterBoxRequest{
		Name: "dev-box", BaseUrl: f.box.URL,
	}))
	wantCode(t, err, connect.CodeAlreadyExists)

	// Invalid base URL is rejected.
	_, err = f.svc.CreateButterBox(f.ctx, connect.NewRequest(&agentsv1.CreateButterBoxRequest{
		Name: "bad", BaseUrl: "not a url",
	}))
	wantCode(t, err, connect.CodeInvalidArgument)

	list, err := f.svc.ListButterBoxes(f.ctx, connect.NewRequest(&agentsv1.ListButterBoxesRequest{}))
	if err != nil || len(list.Msg.GetButterBoxes()) != 1 {
		t.Fatalf("ListButterBoxes = %v, %v", list, err)
	}

	updated, err := f.svc.UpdateButterBox(f.ctx, connect.NewRequest(&agentsv1.UpdateButterBoxRequest{
		Id: created.GetId(), Name: "dev-box-2", BaseUrl: f.box.URL, Enabled: false,
	}))
	if err != nil {
		t.Fatalf("UpdateButterBox: %v", err)
	}
	if updated.Msg.GetButterBox().GetName() != "dev-box-2" || updated.Msg.GetButterBox().GetEnabled() {
		t.Fatalf("update not applied: %+v", updated.Msg.GetButterBox())
	}
	// Update must not clear the stored credential.
	if !updated.Msg.GetButterBox().GetCredentialSet() {
		t.Fatal("credential lost on update")
	}

	if _, err := f.svc.DeleteButterBox(f.ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: created.GetId()})); err != nil {
		t.Fatalf("DeleteButterBox: %v", err)
	}
	_, err = f.svc.GetButterBox(f.ctx, connect.NewRequest(&agentsv1.GetButterBoxRequest{Id: created.GetId()}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestButterBoxTokenWriteOnly(t *testing.T) {
	f := newButterBoxFixture(t)

	created := f.create(t, "dev-box", "")
	if created.GetCredentialSet() {
		t.Fatal("credential_set true without token")
	}

	set, err := f.svc.SetButterBoxToken(f.ctx, connect.NewRequest(&agentsv1.SetButterBoxTokenRequest{
		Id: created.GetId(), Token: "rotated-token",
	}))
	if err != nil || !set.Msg.GetButterBox().GetCredentialSet() {
		t.Fatalf("SetButterBoxToken = %+v, %v", set, err)
	}
	if set.Msg.GetButterBox().GetCredentialUpdatedAt() == nil {
		t.Fatal("credential_updated_at unset after rotation")
	}

	cleared, err := f.svc.SetButterBoxToken(f.ctx, connect.NewRequest(&agentsv1.SetButterBoxTokenRequest{
		Id: created.GetId(), Token: "",
	}))
	if err != nil || cleared.Msg.GetButterBox().GetCredentialSet() {
		t.Fatalf("clear token = %+v, %v", cleared, err)
	}
}

func TestButterBoxTokenRequiresKeyring(t *testing.T) {
	svc := NewButterBoxServiceServer(butterboxmemory.New())
	ctx := workspace.WithID(t.Context(), "ws1")
	_, err := svc.CreateButterBox(ctx, connect.NewRequest(&agentsv1.CreateButterBoxRequest{
		Name: "b", BaseUrl: "https://box.example.com", Token: "tok",
	}))
	wantCode(t, err, connect.CodeFailedPrecondition)
}

func TestButterBoxStatus(t *testing.T) {
	f := newButterBoxFixture(t)
	created := f.create(t, "dev-box", "secret-token")

	status, err := f.svc.GetButterBoxStatus(f.ctx, connect.NewRequest(&agentsv1.GetButterBoxStatusRequest{Id: created.GetId()}))
	if err != nil {
		t.Fatalf("GetButterBoxStatus: %v", err)
	}
	if !status.Msg.GetReachable() || status.Msg.GetActiveSessions() != 2 {
		t.Fatalf("status = %+v", status.Msg)
	}
	// The decrypted token must round-trip to the box as a bearer credential.
	if f.fake.lastAuth != "Bearer secret-token" {
		t.Fatalf("box saw Authorization %q", f.fake.lastAuth)
	}

	// An unreachable box reports status data, not an RPC error.
	f.box.Close()
	status, err = f.svc.GetButterBoxStatus(f.ctx, connect.NewRequest(&agentsv1.GetButterBoxStatusRequest{Id: created.GetId()}))
	if err != nil {
		t.Fatalf("GetButterBoxStatus unreachable: %v", err)
	}
	if status.Msg.GetReachable() || status.Msg.GetError() == "" {
		t.Fatalf("unreachable status = %+v", status.Msg)
	}
}

func TestButterBoxModels(t *testing.T) {
	f := newButterBoxFixture(t)
	created := f.create(t, "dev-box", "secret-token")

	models, err := f.svc.ListButterBoxModels(f.ctx, connect.NewRequest(&agentsv1.ListButterBoxModelsRequest{Id: created.GetId()}))
	if err != nil {
		t.Fatalf("ListButterBoxModels: %v", err)
	}
	got := models.Msg.GetModels()
	if len(got) != 2 || got[0].GetId() != "claude-fable-5" || got[0].GetProvider() != "anthropic" || !got[0].GetReasoning() {
		t.Fatalf("models = %+v", got)
	}
	// The catalog must be requested session-less.
	if f.fake.lastModel.GetSessionId() != "" {
		t.Fatalf("session_id = %q, want empty", f.fake.lastModel.GetSessionId())
	}

	// Unreachable box surfaces as an RPC error here (the caller asked for
	// data, not a health report).
	f.box.Close()
	_, err = f.svc.ListButterBoxModels(f.ctx, connect.NewRequest(&agentsv1.ListButterBoxModelsRequest{Id: created.GetId()}))
	wantCode(t, err, connect.CodeUnavailable)
}

func TestButterBoxRequiresWorkspace(t *testing.T) {
	f := newButterBoxFixture(t)
	_, err := f.svc.ListButterBoxes(t.Context(), connect.NewRequest(&agentsv1.ListButterBoxesRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)
	if err == nil || !strings.Contains(err.Error(), "workspace") {
		t.Fatalf("error should mention workspace: %v", err)
	}
}
