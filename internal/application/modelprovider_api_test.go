package application

import (
	"errors"
	"testing"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestModelProviderAPIRoundTripAndRollback(t *testing.T) {
	store := memory.New()
	svc := NewModelProviderServiceServer(store)
	ctx := testCtx()
	provider := providerWithCapacity("protocols", 32_000)
	provider.OpenaiApi = "responses"
	created, err := svc.CreateModelProvider(ctx, connect.NewRequest(&agentsv1.CreateModelProviderRequest{ModelProvider: provider}))
	if err != nil || created.Msg.GetModelProvider().GetOpenaiApi() != "responses" {
		t.Fatalf("create = (%v, %v), want responses", created, err)
	}
	got, err := svc.GetModelProvider(ctx, connect.NewRequest(&agentsv1.GetModelProviderRequest{Name: provider.Name}))
	if err != nil || got.Msg.GetModelProvider().GetOpenaiApi() != "responses" {
		t.Fatalf("get = (%v, %v), want responses", got, err)
	}
	listed, err := svc.ListModelProviders(ctx, connect.NewRequest(&agentsv1.ListModelProvidersRequest{}))
	if err != nil || len(listed.Msg.GetModelProviders()) != 1 || listed.Msg.GetModelProviders()[0].GetOpenaiApi() != "responses" {
		t.Fatalf("list = (%v, %v), want responses", listed, err)
	}
	provider.OpenaiApi = "chat_completions"
	updated, err := svc.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: provider}))
	if err != nil || updated.Msg.GetModelProvider().GetOpenaiApi() != "chat_completions" {
		t.Fatalf("update = (%v, %v), want chat_completions", updated, err)
	}

	// Protocol edits use the same transactional reload/rollback path as other
	// provider changes; a failed reload must not persist the new protocol.
	runtime := &failOnceRuntime{err: errors.New("reload failed")}
	svc.SetRuntime(runtime)
	provider.OpenaiApi = "responses"
	_, err = svc.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: provider}))
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("failed reload = %v, want Internal", err)
	}
	persisted, err := store.GetModelProvider(t.Context(), wsTest, provider.Name)
	if err != nil || persisted.GetOpenaiApi() != "chat_completions" {
		t.Fatalf("rollback = (%v, %v), want chat_completions", persisted, err)
	}
	if runtime.calls != 2 {
		t.Fatalf("reload calls = %d, want failure and rollback", runtime.calls)
	}
}

func TestModelProviderAPIValidationBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind string
		api  string
	}{
		{name: "unknown", kind: "openai", api: "chat"},
		{name: "case_sensitive", kind: "openai", api: "Responses"},
		{name: "gemini_responses", kind: "gemini", api: "responses"},
		{name: "gemini_chat", kind: "gemini", api: "chat_completions"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := memory.New()
			svc := NewModelProviderServiceServer(store)
			runtime := &reloadTracker{}
			svc.SetRuntime(runtime)
			provider := providerWithCapacity("invalid", 0)
			provider.Type, provider.OpenaiApi = tc.kind, tc.api
			_, err := svc.CreateModelProvider(testCtx(), connect.NewRequest(&agentsv1.CreateModelProviderRequest{ModelProvider: provider}))
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("create error = %v, want InvalidArgument", err)
			}
			all, err := store.ListModelProviders(t.Context(), wsTest)
			if err != nil || len(all) != 0 {
				t.Fatalf("invalid create mutated store: %v, %v", all, err)
			}

			if _, err := store.CreateModelProvider(t.Context(), wsTest, providerWithCapacity("invalid", 0)); err != nil {
				t.Fatal(err)
			}
			_, err = svc.UpdateModelProvider(testCtx(), connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: provider}))
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("update error = %v, want InvalidArgument", err)
			}
			stored, err := store.GetModelProvider(t.Context(), wsTest, "invalid")
			if err != nil || stored.GetType() != "openai" || stored.GetOpenaiApi() != "" {
				t.Fatalf("invalid update mutated store: %v, %v", stored, err)
			}
			if runtime.calls != 0 {
				t.Fatalf("invalid input reloaded runtime %d times", runtime.calls)
			}
		})
	}
}

func TestModelProviderAPIRequiresProvider(t *testing.T) {
	svc := NewModelProviderServiceServer(memory.New())
	_, err := svc.CreateModelProvider(testCtx(), connect.NewRequest(&agentsv1.CreateModelProviderRequest{}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("missing provider create = %v, want InvalidArgument", err)
	}
	_, err = svc.UpdateModelProvider(testCtx(), connect.NewRequest(&agentsv1.UpdateModelProviderRequest{}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("missing provider update = %v, want InvalidArgument", err)
	}
}

func TestModelProviderAPILegacyDefault(t *testing.T) {
	for _, kind := range []string{"openai", "gemini"} {
		t.Run(kind, func(t *testing.T) {
			svc := NewModelProviderServiceServer(memory.New())
			provider := providerWithCapacity("legacy", 0)
			provider.Type = kind
			res, err := svc.CreateModelProvider(testCtx(), connect.NewRequest(&agentsv1.CreateModelProviderRequest{ModelProvider: provider}))
			if err != nil || res.Msg.GetModelProvider().GetOpenaiApi() != "" {
				t.Fatalf("omitted API = (%v, %v), want unchanged empty value", res, err)
			}
		})
	}
}
