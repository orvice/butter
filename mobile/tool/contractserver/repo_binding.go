package main

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/gitprovider"
	agentcontentmemory "go.orx.me/apps/butter/internal/repo/agentcontent/memory"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	githostmemory "go.orx.me/apps/butter/internal/repo/githost/memory"
	repobindingmemory "go.orx.me/apps/butter/internal/repo/repobinding/memory"
	repocachememory "go.orx.me/apps/butter/internal/repo/repocache/memory"
	"go.orx.me/apps/butter/internal/repo/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const (
	repoBindingWorkspace       = "repo"
	repoBindingMemberWorkspace = "repo-member"
	repoBindingGitHostID       = "contract-git-host"
	repoBindingEncryptionKey   = "0123456789abcdef0123456789abcdef"
)

type repoBindingFixture struct {
	svc       *application.RepoBindingServiceServer
	hostRepo  *githostmemory.Store
	binding   *repobindingmemory.Store
	fakeGit   *fakeGitClient
	agentRepo *configmemory.Store
}

func newRepoBindingFixture(
	ctx context.Context,
	wsRepo workspace.Repository,
	agentRepo *configmemory.Store,
) (*repoBindingFixture, error) {
	hostRepo := githostmemory.New()
	if _, err := hostRepo.Create(ctx, &agentsv1.GitHost{
		Id: repoBindingGitHostID, Name: "Contract GitHub",
		Kind:       agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB,
		ApiBaseUrl: "https://api.github.com",
	}); err != nil {
		return nil, err
	}
	bindingRepo := repobindingmemory.New()
	cacheRepo := repocachememory.New()
	contentRepo := agentcontentmemory.New()
	fake := newFakeGitClient()
	fake.materialize = true

	svc := application.NewRepoBindingServiceServer(bindingRepo, hostRepo)
	svc.SetWorkspaceRepo(wsRepo)
	svc.SetAgentRepo(agentRepo)
	svc.SetCacheRepo(cacheRepo)
	svc.SetContentRepo(contentRepo)
	svc.SetEncryptionKeyProvider(func() string { return repoBindingEncryptionKey })
	svc.SetProviderClientFactory(func(cfg gitprovider.Config) (gitprovider.Client, error) {
		if cfg.Token == "deny-read" {
			f := newFakeGitClient()
			f.repoErr = gitprovider.ErrUnauthorized
			return f, nil
		}
		return fake, nil
	})

	if _, err := agentRepo.CreateAgent(ctx, repoBindingWorkspace, &agentsv1.Agent{
		AgentId: "repo-agent", Name: "repo-agent",
		Description: "DB description",
		Config:      &agentsv1.AgentConfig{Instruction: "DB prompt"},
	}); err != nil {
		return nil, err
	}

	return &repoBindingFixture{
		svc: svc, hostRepo: hostRepo, binding: bindingRepo, fakeGit: fake, agentRepo: agentRepo,
	}, nil
}

func (fx *repoBindingFixture) register(r *gin.Engine, opts []connect.HandlerOption) {
	_, handler := agentsv1connect.NewWorkspaceRepoBindingServiceHandler(fx.svc, opts...)
	gitSvc := application.NewGitHostServiceServer(fx.hostRepo)
	_, gitHandler := agentsv1connect.NewGitHostServiceHandler(gitSvc, opts...)
	for _, procedure := range []string{
		agentsv1connect.WorkspaceRepoBindingServiceGetWorkspaceRepoBindingProcedure,
		agentsv1connect.WorkspaceRepoBindingServicePutWorkspaceRepoBindingProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceDeleteWorkspaceRepoBindingProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceOnboardWorkspaceRepositoryProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceSetWorkspaceRepoBindingCredentialProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceValidateWorkspaceRepoBindingProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceSyncWorkspaceRepositoryProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceGetRepositorySyncStatusProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceListRepositoryEntriesProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceGetRepositoryFileProcedure,
		agentsv1connect.WorkspaceRepoBindingServicePublishWorkspaceRepositoryProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceConfigureWebhookSecretProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceAcceptRepositoryBaselineProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceCommitAgentContentProcedure,
		agentsv1connect.WorkspaceRepoBindingServiceRollbackAgentContentProcedure,
		agentsv1connect.WorkspaceRepoBindingServicePurgeAgentContentProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", handler)))
	}
	for _, procedure := range []string{
		agentsv1connect.GitHostServiceListGitHostsProcedure,
		agentsv1connect.GitHostServiceGetGitHostProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", gitHandler)))
	}
}
