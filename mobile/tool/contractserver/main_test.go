package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// providerFixture serves the production Connect handlers and middleware that
// the Flutter app talks to, with only persistence replaced by memory.
type providerFixture struct {
	t     *testing.T
	token string
	url   string
}

func newProviderFixture(t *testing.T) *providerFixture {
	t.Helper()
	handler, err := fixture(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	auth := agentsv1connect.NewAuthServiceClient(server.Client(), server.URL+"/api")
	login, err := auth.Login(context.Background(), connect.NewRequest(&agentsv1.LoginRequest{
		Username: "mobile-test", Password: "contract-test-password",
	}))
	if err != nil {
		t.Fatal(err)
	}
	return &providerFixture{t: t, token: login.Msg.GetToken(), url: server.URL}
}

// client acts in a Workspace; an empty ID sends no X-Workspace-ID header.
func (f *providerFixture) client(workspaceID string) agentsv1connect.ModelProviderServiceClient {
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+f.token)
			if workspaceID != "" {
				req.Header().Set("X-Workspace-ID", workspaceID)
			}
			return next(ctx, req)
		}
	})
	return agentsv1connect.NewModelProviderServiceClient(http.DefaultClient, f.url+"/api", connect.WithInterceptors(interceptor))
}

func wantCode(t *testing.T, err error, code connect.Code) *connect.Error {
	t.Helper()
	var got *connect.Error
	if err == nil {
		t.Fatalf("expected %v, got success", code)
	}
	if !errors.As(err, &got) || got.Code() != code {
		t.Fatalf("expected %v, got %v", code, err)
	}
	return got
}

// withUnknown appends an unknown varint field the way a newer server would.
func withUnknown(m proto.Message, number int32, value uint64) {
	raw := m.ProtoReflect().GetUnknown()
	raw = protowire.AppendTag(raw, protowire.Number(number), protowire.VarintType)
	raw = protowire.AppendVarint(raw, value)
	m.ProtoReflect().SetUnknown(raw)
}

func TestModelProviderCRUDRoundTripsEveryField(t *testing.T) {
	f := newProviderFixture(t)
	ctx := context.Background()
	alpha := f.client("alpha")

	list, err := alpha.ListModelProviders(ctx, connect.NewRequest(&agentsv1.ListModelProvidersRequest{}))
	if err != nil || len(list.Msg.GetModelProviders()) != 0 {
		t.Fatalf("initial list = %v, %v", list, err)
	}

	model := &agentsv1.ModelConfig{Name: "gpt-4o", Alias: "fast", ContextWindowTokens: 4294967295}
	withUnknown(model, 15, 3)
	draft := &agentsv1.ModelProvider{
		Name: "openai", Type: "openai", ApiKey: "sk-secret", BaseUrl: "https://llm.example/v1",
		Models: []*agentsv1.ModelConfig{model, {Name: "gpt-mini"}},
	}
	withUnknown(draft, 99, 7)
	created, err := alpha.CreateModelProvider(ctx, connect.NewRequest(&agentsv1.CreateModelProviderRequest{ModelProvider: draft}))
	if err != nil {
		t.Fatal(err)
	}
	got := created.Msg.GetModelProvider()
	if got.GetWorkspaceId() != "alpha" || got.GetApiKey() != "sk-secret" ||
		got.GetModels()[0].GetContextWindowTokens() != 4294967295 || got.GetModels()[1].GetAlias() != "" {
		t.Fatalf("created = %v", got)
	}
	if len(got.ProtoReflect().GetUnknown()) == 0 || len(got.GetModels()[0].ProtoReflect().GetUnknown()) == 0 {
		t.Fatal("unknown fields were dropped on create")
	}

	// Update replaces the whole object: what is sent is what is stored, so a
	// client that wants to keep the key must send it back.
	keep := proto.Clone(got).(*agentsv1.ModelProvider)
	keep.BaseUrl = ""
	updated, err := alpha.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: keep}))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Msg.GetModelProvider().GetApiKey() != "sk-secret" || updated.Msg.GetModelProvider().GetBaseUrl() != "" ||
		len(updated.Msg.GetModelProvider().GetModels()) != 2 {
		t.Fatalf("keep update = %v", updated.Msg.GetModelProvider())
	}
	cleared := proto.Clone(keep).(*agentsv1.ModelProvider)
	cleared.ApiKey = ""
	updated, err = alpha.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: cleared}))
	if err != nil || updated.Msg.GetModelProvider().GetApiKey() != "" {
		t.Fatalf("clear update = %v, %v", updated, err)
	}
	// An update that omits the key clears it; there is no write-only patch.
	read, err := alpha.GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: "openai"}))
	if err != nil || read.Msg.GetModelProvider().GetApiKey() != "" || len(read.Msg.GetModelProvider().ProtoReflect().GetUnknown()) == 0 {
		t.Fatalf("get after clear = %v, %v", read, err)
	}

	if _, err := alpha.DeleteModelProvider(ctx, connect.NewRequest(&agentsv1.DeleteModelProviderRequest{Name: "openai"})); err != nil {
		t.Fatal(err)
	}
	_, err = alpha.GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: "openai"}))
	wantCode(t, err, connect.CodeNotFound)
	_, err = alpha.DeleteModelProvider(ctx, connect.NewRequest(&agentsv1.DeleteModelProviderRequest{Name: "openai"}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestModelProviderWorkspaceHeaderIsolatesAndGuards(t *testing.T) {
	f := newProviderFixture(t)
	ctx := context.Background()
	draft := func() *connect.Request[agentsv1.CreateModelProviderRequest] {
		return connect.NewRequest(&agentsv1.CreateModelProviderRequest{ModelProvider: &agentsv1.ModelProvider{
			Name: "shared-name", Type: "gemini", Models: []*agentsv1.ModelConfig{{Name: "m"}},
		}})
	}
	if _, err := f.client("alpha").CreateModelProvider(ctx, draft()); err != nil {
		t.Fatal(err)
	}
	_, err := f.client("alpha").CreateModelProvider(ctx, draft())
	wantCode(t, err, connect.CodeAlreadyExists)

	// The same name lives independently in another Workspace.
	if _, err := f.client("beta").CreateModelProvider(ctx, draft()); err != nil {
		t.Fatal(err)
	}
	beta := f.client("beta")
	update := connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: &agentsv1.ModelProvider{
		Name: "shared-name", Type: "openai", Models: []*agentsv1.ModelConfig{{Name: "m"}},
	}})
	if _, err := beta.UpdateModelProvider(ctx, update); err != nil {
		t.Fatal(err)
	}
	alphaRead, err := f.client("alpha").GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: "shared-name"}))
	if err != nil || alphaRead.Msg.GetModelProvider().GetType() != "gemini" || alphaRead.Msg.GetModelProvider().GetWorkspaceId() != "alpha" {
		t.Fatalf("alpha provider changed by a beta update: %v, %v", alphaRead, err)
	}

	// No header: the service demands a Workspace. An unknown or foreign one is
	// refused by the auth middleware before the service runs.
	_, err = f.client("").ListModelProviders(ctx, connect.NewRequest(&agentsv1.ListModelProvidersRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)
	_, err = f.client("forbidden").ListModelProviders(ctx, connect.NewRequest(&agentsv1.ListModelProvidersRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)

	// A member (not an owner) may manage providers: the service has no role check.
	if _, err := beta.DeleteModelProvider(ctx, connect.NewRequest(&agentsv1.DeleteModelProviderRequest{Name: "shared-name"})); err != nil {
		t.Fatal(err)
	}
}

func TestModelProviderValidation(t *testing.T) {
	f := newProviderFixture(t)
	ctx := context.Background()
	alpha := f.client("alpha")
	for _, bad := range []string{"not a url", "ftp://host/x", "/relative"} {
		_, err := alpha.CreateModelProvider(ctx, connect.NewRequest(&agentsv1.CreateModelProviderRequest{
			ModelProvider: &agentsv1.ModelProvider{Name: "bad", Type: "openai", BaseUrl: bad},
		}))
		err2 := wantCode(t, err, connect.CodeInvalidArgument)
		if strings.Contains(err2.Message(), "sk-") {
			t.Fatalf("error leaks input: %v", err2)
		}
	}
	// The failed create left nothing behind, and an unknown provider cannot be updated.
	_, err := alpha.GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: "bad"}))
	wantCode(t, err, connect.CodeNotFound)
	_, err = alpha.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{
		ModelProvider: &agentsv1.ModelProvider{Name: "missing", Type: "openai"},
	}))
	wantCode(t, err, connect.CodeNotFound)

	created, err := alpha.CreateModelProvider(ctx, connect.NewRequest(&agentsv1.CreateModelProviderRequest{
		ModelProvider: &agentsv1.ModelProvider{Name: "ok", Type: "openai", ApiKey: "sk-keep", BaseUrl: "https://ok.example"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	bad := proto.Clone(created.Msg.GetModelProvider()).(*agentsv1.ModelProvider)
	bad.BaseUrl = "ftp://nope"
	_, err = alpha.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: bad}))
	wantCode(t, err, connect.CodeInvalidArgument)
	read, err := alpha.GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: "ok"}))
	if err != nil || read.Msg.GetModelProvider().GetBaseUrl() != "https://ok.example" || read.Msg.GetModelProvider().GetApiKey() != "sk-keep" {
		t.Fatalf("a rejected update changed the provider: %v, %v", read, err)
	}
}

func TestModelProviderTelegramReferencesBlockRemoval(t *testing.T) {
	f := newProviderFixture(t)
	ctx := context.Background()
	alpha, beta := f.client("alpha"), f.client("beta")
	provider := func() *agentsv1.ModelProvider {
		return &agentsv1.ModelProvider{Name: "routed", Type: "openai", Models: []*agentsv1.ModelConfig{
			{Name: "real-a", Alias: guardedModelAlias},
			{Name: "real-b", Alias: guardedSelectableAlias},
			{Name: "free"},
		}}
	}
	for _, c := range []agentsv1connect.ModelProviderServiceClient{alpha, beta} {
		if _, err := c.CreateModelProvider(ctx, connect.NewRequest(&agentsv1.CreateModelProviderRequest{ModelProvider: provider()})); err != nil {
			t.Fatal(err)
		}
	}

	// Dropping a referenced model is refused, naming the Destination.
	drop := provider()
	drop.Models = drop.Models[1:]
	_, err := alpha.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: drop}))
	e := wantCode(t, err, connect.CodeFailedPrecondition)
	for _, want := range []string{"referenced by telegram destinations", guardedModelAlias, guardedDestinationID} {
		if !strings.Contains(e.Message(), want) {
			t.Fatalf("message %q lacks %q", e.Message(), want)
		}
	}
	// Deleting the provider is refused too, and nothing changed.
	_, err = alpha.DeleteModelProvider(ctx, connect.NewRequest(&agentsv1.DeleteModelProviderRequest{Name: "routed"}))
	e = wantCode(t, err, connect.CodeFailedPrecondition)
	if !strings.Contains(e.Message(), guardedSelectableAlias) {
		t.Fatalf("message %q lacks the selectable alias", e.Message())
	}
	read, err := alpha.GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: "routed"}))
	if err != nil || len(read.Msg.GetModelProvider().GetModels()) != 3 {
		t.Fatalf("a refused change was applied: %v, %v", read, err)
	}

	// Renaming an unreferenced model, and keeping the referenced ones, is fine.
	keep := provider()
	keep.Models[2].Name = "renamed-free"
	if _, err := alpha.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: keep})); err != nil {
		t.Fatal(err)
	}
	// References belong to a Workspace: beta has no Destination, so the same
	// alias is removable there.
	if _, err := beta.DeleteModelProvider(ctx, connect.NewRequest(&agentsv1.DeleteModelProviderRequest{Name: "routed"})); err != nil {
		t.Fatal(err)
	}
}
