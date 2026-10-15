package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"connectrpc.com/connect"
	cursorv1 "github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1/cursorv1connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/application"
	butterboxmemory "go.orx.me/apps/butter/internal/repo/butterbox/memory"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/runtime/cursorbox"
	"go.orx.me/apps/butter/internal/runtime/pibox"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// The ButterBox fixture: the production ButterBoxService, AgentService leaf
// validation and the real pibox and cursorbox bridges, with only the Box
// itself replaced by an in-process test double that serves the real generated
// PiService and CursorService handlers. It is not a production Box and not a
// device run. It lives in its own Workspace so the contracts that count other
// Workspaces' agents and boxes are untouched.
//
// The double's behavior is chosen by the base URL path the box is registered
// with, so a contract test picks a failure by choosing a URL:
//
//	/fake-box/ok        works; requires the access token boxToken
//	/fake-box/capacity  every Pi and Cursor session is refused as at capacity
//	/fake-box/nokey     the Cursor bridge answers with its missing-API-key reason
//
// Any other address (for example a closed port) is simply unreachable.
const (
	boxWorkspace = "boxes"
	// boxToken is the access token the /ok box requires; it is public test
	// data for this loopback fixture only.
	boxToken = "box-secret-token"
)

// fakeBox is one profile of the double. Replies carry the session, working
// directory and model the Butter server asked for, so a test observes
// continuity and repointing from the Agent's own answers.
type fakeBox struct {
	profile string

	mu       sync.Mutex
	next     int
	piSess   map[string]*piv1.CreateSessionRequest
	curSess  map[string]*cursorv1.CreateSessionRequest
	turns    map[string]string
	piModels []*piv1.Model
}

func newFakeBox(profile string) *fakeBox {
	return &fakeBox{
		profile: profile,
		piSess:  map[string]*piv1.CreateSessionRequest{},
		curSess: map[string]*cursorv1.CreateSessionRequest{},
		turns:   map[string]string{},
		piModels: []*piv1.Model{
			{Id: "claude-fable-5", Provider: "anthropic", Name: "Claude Fable 5", Reasoning: true},
			{Id: "gpt-5.6", Provider: "openai", Name: "GPT-5.6"},
		},
	}
}

func (f *fakeBox) authorize(h http.Header) error {
	if f.profile != "ok" {
		return nil
	}
	if h.Get("Authorization") != "Bearer "+boxToken {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("box access token missing or wrong"))
	}
	return nil
}

func (f *fakeBox) nextID(prefix string) string {
	f.next++
	return fmt.Sprintf("%s-%s-%d", prefix, f.profile, f.next)
}

func (f *fakeBox) atCapacity() error {
	if f.profile == "capacity" {
		return connect.NewError(connect.CodeResourceExhausted, errors.New("box session limit reached"))
	}
	return nil
}

func noCursorKey() error {
	err := connect.NewError(connect.CodeUnauthenticated, errors.New("cursor API key is missing or invalid"))
	if detail, derr := connect.NewErrorDetail(&errdetails.ErrorInfo{Reason: "CURSOR_API_KEY_MISSING_OR_INVALID"}); derr == nil {
		err.AddDetail(detail)
	}
	return err
}

func (f piBox) ListSessions(_ context.Context, req *connect.Request[piv1.ListSessionsRequest]) (*connect.Response[piv1.ListSessionsResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	sessions := make([]*piv1.Session, 0, len(f.piSess))
	for id := range f.piSess {
		sessions = append(sessions, &piv1.Session{Id: id})
	}
	return connect.NewResponse(&piv1.ListSessionsResponse{Sessions: sessions}), nil
}

func (f piBox) GetAvailableModels(_ context.Context, req *connect.Request[piv1.GetAvailableModelsRequest]) (*connect.Response[piv1.GetAvailableModelsResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	if err := f.atCapacity(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&piv1.GetAvailableModelsResponse{Models: f.piModels}), nil
}

func (f piBox) CreateSession(_ context.Context, req *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	if err := f.atCapacity(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id := f.nextID("pi")
	f.piSess[id] = proto.Clone(req.Msg).(*piv1.CreateSessionRequest)
	return connect.NewResponse(&piv1.CreateSessionResponse{Session: &piv1.Session{Id: id, Cwd: req.Msg.GetCwd()}}), nil
}

func (f piBox) SubmitMessage(_ context.Context, req *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	created, ok := f.piSess[req.Msg.GetSessionId()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown pi session"))
	}
	cursor := fmt.Sprintf("cursor-%d", len(f.turns)+1)
	f.turns[cursor] = fmt.Sprintf("pi[session=%s,cwd=%s,model=%s/%s,thinking=%s](%s)",
		req.Msg.GetSessionId(), created.GetCwd(), created.GetProvider(), created.GetModel(),
		created.GetThinkingLevel(), req.Msg.GetMessage())
	return connect.NewResponse(&piv1.SubmitMessageResponse{TurnCursor: cursor}), nil
}

func (f piBox) GetTurn(_ context.Context, req *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	text, ok := f.turns[req.Msg.GetTurnCursor()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown turn"))
	}
	return connect.NewResponse(&piv1.GetTurnResponse{Result: &piv1.TurnResult{Text: text, StopReason: "stop"}}), nil
}

func (f piBox) AbortSession(_ context.Context, _ *connect.Request[piv1.AbortSessionRequest]) (*connect.Response[piv1.AbortSessionResponse], error) {
	return connect.NewResponse(&piv1.AbortSessionResponse{}), nil
}

func (f cursorBox) ListModels(_ context.Context, req *connect.Request[cursorv1.ListModelsRequest]) (*connect.Response[cursorv1.ListModelsResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	if f.profile == "nokey" {
		return nil, noCursorKey()
	}
	return connect.NewResponse(&cursorv1.ListModelsResponse{Models: []*cursorv1.Model{
		{Id: "composer-2.5", Name: "Composer 2.5"},
		{Id: "auto-smart", Name: "Auto (smart)"},
	}}), nil
}

func (f cursorBox) CreateSession(_ context.Context, req *connect.Request[cursorv1.CreateSessionRequest]) (*connect.Response[cursorv1.CreateSessionResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	if f.profile == "nokey" {
		return nil, noCursorKey()
	}
	if err := f.atCapacity(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id := f.nextID("cursor")
	f.curSess[id] = proto.Clone(req.Msg).(*cursorv1.CreateSessionRequest)
	return connect.NewResponse(&cursorv1.CreateSessionResponse{SessionId: id}), nil
}

func (f cursorBox) SendMessage(_ context.Context, req *connect.Request[cursorv1.SendMessageRequest]) (*connect.Response[cursorv1.SendMessageResponse], error) {
	if err := f.authorize(req.Header()); err != nil {
		return nil, err
	}
	if f.profile == "nokey" {
		return nil, noCursorKey()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	created, ok := f.curSess[req.Msg.GetSessionId()]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown cursor session"))
	}
	return connect.NewResponse(&cursorv1.SendMessageResponse{Text: fmt.Sprintf(
		"cursor[session=%s,cwd=%s,model=%s,mode=%s](%s)",
		req.Msg.GetSessionId(), created.GetCwd(), created.GetModel(), created.GetMode(), req.Msg.GetMessage())}), nil
}

func (f cursorBox) AbortSession(_ context.Context, _ *connect.Request[cursorv1.AbortSessionRequest]) (*connect.Response[cursorv1.AbortSessionResponse], error) {
	return connect.NewResponse(&cursorv1.AbortSessionResponse{}), nil
}

type piBox struct {
	*fakeBox
	piv1connect.UnimplementedPiServiceHandler
}

type cursorBox struct {
	*fakeBox
	cursorv1connect.UnimplementedCursorServiceHandler
}

// boxDouble serves the double on the fixture's own listener under
// /fake-box/, so a box registered at <fixture URL>/fake-box/<profile> is
// reached by the Butter server over real HTTP and the contract test needs no
// second address.
func boxDouble() http.Handler {
	mux := http.NewServeMux()
	for _, profile := range []string{"ok", "capacity", "nokey"} {
		box := newFakeBox(profile)
		sub := http.NewServeMux()
		piPath, piHandler := piv1connect.NewPiServiceHandler(piBox{fakeBox: box})
		sub.Handle(piPath, piHandler)
		cursorPath, cursorHandler := cursorv1connect.NewCursorServiceHandler(cursorBox{fakeBox: box})
		sub.Handle(cursorPath, cursorHandler)
		prefix := "/fake-box/" + profile
		mux.Handle(prefix+"/", http.StripPrefix(prefix, sub))
	}
	return mux
}

// boxFixture is the production ButterBox wiring over in-memory storage.
type boxFixture struct {
	repo     *butterboxmemory.Store
	keyring  *secretbox.Keyring
	svc      *application.ButterBoxServiceServer
	builders *internalagent.BoxAgentBuilders
}

func newBoxFixture(agentRepo *configmemory.Store) *boxFixture {
	repo := butterboxmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	svc := application.NewButterBoxServiceServer(repo)
	svc.SetKeyring(keyring)
	svc.SetAgentRepo(agentRepo)
	return &boxFixture{
		repo: repo, keyring: keyring, svc: svc,
		builders: &internalagent.BoxAgentBuilders{
			Pi:     pibox.AgentBuilder(pibox.NewFactory(repo, keyring)),
			Cursor: cursorbox.AgentBuilder(cursorbox.NewFactory(repo, keyring)),
		},
	}
}

// reloader is the production ConfigRuntime contract over the fixture: a saved
// Agent is rebuilt into the runner, as the real server does after a write, so
// a leaf created through AgentService can run at once. Only the Workspaces the
// runner serves are loaded; the others hold agents with no runnable model.
type reloader struct {
	runner     *runner.Service
	agents     *configmemory.Store
	workspaces []string
	providers  []agentsv1.ModelProvider
}

func (r reloader) ReloadRunner(ctx context.Context) error {
	var all []agentsv1.Agent
	for _, ws := range r.workspaces {
		list, err := r.agents.ListAgents(ctx, ws)
		if err != nil {
			return err
		}
		for _, stored := range list {
			all = append(all, agentsv1.Agent{})
			proto.Merge(&all[len(all)-1], stored)
			all[len(all)-1].WorkspaceId = ws
		}
	}
	return r.runner.ReloadProtoAgents(ctx, all, r.providers, nil, nil)
}

func (reloader) ReloadChannels(context.Context) error { return nil }

var _ application.ConfigRuntime = reloader{}
