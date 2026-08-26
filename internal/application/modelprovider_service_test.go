package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type failOnceRuntime struct {
	calls int
	err   error
}

func (r *failOnceRuntime) ReloadRunner(context.Context) error {
	r.calls++
	if r.calls == 1 {
		return r.err
	}
	return nil
}

func (r *failOnceRuntime) ReloadChannels(context.Context) error { return nil }

func providerWithCapacity(name string, capacity uint32) *agentsv1.ModelProvider {
	return &agentsv1.ModelProvider{
		Name: name,
		Type: "openai",
		Models: []*agentsv1.ModelConfig{{
			Name:                "custom-model",
			Alias:               "custom",
			ContextWindowTokens: capacity,
		}},
	}
}

func requireProviderCapacity(t *testing.T, provider *agentsv1.ModelProvider, want uint32) {
	t.Helper()
	if provider == nil || len(provider.GetModels()) != 1 {
		t.Fatalf("provider = %v, want one model", provider)
	}
	if got := provider.GetModels()[0].GetContextWindowTokens(); got != want {
		t.Fatalf("context_window_tokens = %d, want %d", got, want)
	}
}

func TestModelProviderServiceContextCapacityRoundTripsAcrossRPCs(t *testing.T) {
	store := memory.New()
	svc := NewModelProviderServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)
	ctx := testCtx()

	created, err := svc.CreateModelProvider(ctx, connect.NewRequest(&agentsv1.CreateModelProviderRequest{
		ModelProvider: providerWithCapacity("custom", 64_000),
	}))
	if err != nil {
		t.Fatalf("CreateModelProvider: %v", err)
	}
	requireProviderCapacity(t, created.Msg.GetModelProvider(), 64_000)

	got, err := svc.GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: "custom"}))
	if err != nil {
		t.Fatalf("GetModelProvider: %v", err)
	}
	requireProviderCapacity(t, got.Msg.GetModelProvider(), 64_000)

	listed, err := svc.ListModelProviders(ctx, connect.NewRequest(&agentsv1.ListModelProvidersRequest{}))
	if err != nil {
		t.Fatalf("ListModelProviders: %v", err)
	}
	if len(listed.Msg.GetModelProviders()) != 1 {
		t.Fatalf("ListModelProviders returned %d providers, want 1", len(listed.Msg.GetModelProviders()))
	}
	requireProviderCapacity(t, listed.Msg.GetModelProviders()[0], 64_000)

	updatedProvider := proto.Clone(got.Msg.GetModelProvider()).(*agentsv1.ModelProvider)
	updatedProvider.Models[0].ContextWindowTokens = 96_000
	updated, err := svc.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{
		ModelProvider: updatedProvider,
	}))
	if err != nil {
		t.Fatalf("UpdateModelProvider: %v", err)
	}
	requireProviderCapacity(t, updated.Msg.GetModelProvider(), 96_000)
	if runtime.calls != 2 {
		t.Fatalf("runtime reloads = %d, want create and update reloads", runtime.calls)
	}
}

func TestModelProviderServiceReloadFailureRollsBackCapacityUpdate(t *testing.T) {
	store := memory.New()
	if _, err := store.CreateModelProvider(context.Background(), wsTest, providerWithCapacity("custom", 32_000)); err != nil {
		t.Fatalf("seed provider: %v", err)
	}

	runtime := &failOnceRuntime{err: errors.New(`model "custom-model" has conflicting context_window_tokens values: 32000, 64000`)}
	svc := NewModelProviderServiceServer(store)
	svc.SetRuntime(runtime)

	_, err := svc.UpdateModelProvider(testCtx(), connect.NewRequest(&agentsv1.UpdateModelProviderRequest{
		ModelProvider: providerWithCapacity("custom", 64_000),
	}))
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("UpdateModelProvider error = %v, want Internal", err)
	}
	if runtime.calls != 2 {
		t.Fatalf("runtime reloads = %d, want failed update reload plus successful rollback reload", runtime.calls)
	}

	stored, getErr := store.GetModelProvider(context.Background(), wsTest, "custom")
	if getErr != nil {
		t.Fatalf("GetModelProvider after rollback: %v", getErr)
	}
	requireProviderCapacity(t, stored, 32_000)
}
