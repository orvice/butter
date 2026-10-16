package main

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/application"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	memoryconfigmem "go.orx.me/apps/butter/internal/repo/memoryconfig/memory"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/mem0memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	agentsv1connect "go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const (
	memoryWorkspace       = "memory"
	memoryMemberWorkspace = "memory-member"
)

type memoryFixture struct {
	store  *mem0Store
	repo   *memoryconfigmem.Store
	config *application.WorkspaceMemoryConfigServiceServer
	memSvc *application.WorkspaceMemoryServiceServer
}

func newMemoryFixture(wsRepo workspacerepo.Repository) *memoryFixture {
	store := newMem0Store()
	store.seed(
		seedContractMemory("w-new", "team uses pnpm", "ws:"+memoryWorkspace, "", "2026-09-20T10:00:00.5+00:00"),
		seedContractMemory("w-old", "team deploys on friday", "ws:"+memoryWorkspace, "", "2026-09-01T10:00:00+00:00"),
		seedContractMemory("a-1", "answer tersely", "", "ws:"+memoryWorkspace+":agent:mem-agent", "2026-09-10T10:00:00+00:00"),
		seedContractMemory("foreign-w", "other tenant secret", "ws:foreign", "", "2026-09-15T10:00:00+00:00"),
		seedContractMemory("mm-1", "member pool note", "ws:"+memoryMemberWorkspace, "", "2026-09-10T10:00:00+00:00"),
	)
	repo := memoryconfigmem.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	configSvc := application.NewWorkspaceMemoryConfigServiceServer(repo)
	configSvc.SetKeyring(keyring)
	configSvc.SetWorkspaceRepo(wsRepo)
	configSvc.SetHTTPClient(store.srv.Client())

	mem0Svc := mem0memory.New(memoryconn.NewResolver(repo, keyring))
	mem0Svc.HTTPClient = store.srv.Client()
	memSvc := application.NewWorkspaceMemoryServiceServer(mem0Svc)
	memSvc.SetWorkspaceRepo(wsRepo)

	fx := &memoryFixture{store: store, repo: repo, config: configSvc, memSvc: memSvc}
	fx.seedWorkspaceConfig(context.Background(), memoryWorkspace, keyring)
	fx.seedWorkspaceConfig(context.Background(), memoryMemberWorkspace, keyring)
	return fx
}

func (fx *memoryFixture) seedWorkspaceConfig(ctx context.Context, workspaceID string, keyring *secretbox.Keyring) {
	ct, kid, err := keyring.Encrypt(ctx, []byte(contractMem0Key))
	if err != nil {
		panic(err)
	}
	if _, err := fx.repo.Put(ctx, workspaceID, &agentsv1.WorkspaceMemoryConfig{
		BaseUrl: fx.store.baseURL(), Enabled: true,
	}, &memoryconfigrepo.Credential{Ciphertext: ct, KeyID: kid}); err != nil {
		panic(err)
	}
}

func (fx *memoryFixture) register(r *gin.Engine, opts []connect.HandlerOption) {
	_, configHandler := agentsv1connect.NewWorkspaceMemoryConfigServiceHandler(fx.config, opts...)
	for _, procedure := range []string{
		agentsv1connect.WorkspaceMemoryConfigServiceGetWorkspaceMemoryConfigProcedure,
		agentsv1connect.WorkspaceMemoryConfigServicePutWorkspaceMemoryConfigProcedure,
		agentsv1connect.WorkspaceMemoryConfigServiceDeleteWorkspaceMemoryConfigProcedure,
		agentsv1connect.WorkspaceMemoryConfigServiceTestWorkspaceMemoryConnectionProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", configHandler)))
	}
	_, memHandler := agentsv1connect.NewWorkspaceMemoryServiceHandler(fx.memSvc, opts...)
	for _, procedure := range []string{
		agentsv1connect.WorkspaceMemoryServiceListWorkspaceMemoriesProcedure,
		agentsv1connect.WorkspaceMemoryServiceSearchWorkspaceMemoriesProcedure,
		agentsv1connect.WorkspaceMemoryServiceDeleteWorkspaceMemoryProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", memHandler)))
	}
}

func init() {
	registerFixtureWorkspaces(
		fixtureWorkspace{ID: memoryWorkspace, Owner: true, NoSeedAgent: true},
		fixtureWorkspace{ID: memoryMemberWorkspace, NoSeedAgent: true},
	)
}
