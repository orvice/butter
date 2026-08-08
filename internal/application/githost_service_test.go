package application

// Service-level tests for GitHostService (issue #214): global-admin-only
// mutations, open reads, and API base URL validation (the SSRF guard).

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/auth"
	githostmemory "go.orx.me/apps/butter/internal/repo/githost/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func newGitHostService() *GitHostServiceServer {
	return NewGitHostServiceServer(githostmemory.New())
}

func validHost() *agentsv1.GitHost {
	return &agentsv1.GitHost{
		Name:       "GitLab self-hosted",
		Kind:       agentsv1.GitHostKind_GIT_HOST_KIND_GITLAB,
		ApiBaseUrl: "https://gitlab.example.com/api/v4",
		WebBaseUrl: "https://gitlab.example.com",
	}
}

func TestGitHostAdminOnlyMutations(t *testing.T) {
	svc := newGitHostService()
	user := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "u1", Role: "user"}, nil)
	admin := auth.WithAdmin(context.Background())

	_, err := svc.CreateGitHost(user, connect.NewRequest(&agentsv1.CreateGitHostRequest{Host: validHost()}))
	wantCode(t, err, connect.CodePermissionDenied)

	created, err := svc.CreateGitHost(admin, connect.NewRequest(&agentsv1.CreateGitHostRequest{Host: validHost()}))
	if err != nil {
		t.Fatalf("admin Create: %v", err)
	}
	id := created.Msg.GetHost().GetId()
	if id == "" {
		t.Fatal("expected server-assigned host id")
	}

	mod := validHost()
	mod.Id = id
	mod.Name = "Renamed"
	_, err = svc.UpdateGitHost(user, connect.NewRequest(&agentsv1.UpdateGitHostRequest{Host: mod}))
	wantCode(t, err, connect.CodePermissionDenied)
	if _, err := svc.UpdateGitHost(admin, connect.NewRequest(&agentsv1.UpdateGitHostRequest{Host: mod})); err != nil {
		t.Fatalf("admin Update: %v", err)
	}

	_, err = svc.DeleteGitHost(user, connect.NewRequest(&agentsv1.DeleteGitHostRequest{Id: id}))
	wantCode(t, err, connect.CodePermissionDenied)

	// Reads are open to any authenticated caller.
	list, err := svc.ListGitHosts(user, connect.NewRequest(&agentsv1.ListGitHostsRequest{}))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Msg.GetHosts()) != 1 || list.Msg.GetHosts()[0].GetName() != "Renamed" {
		t.Fatalf("unexpected hosts: %v", list.Msg.GetHosts())
	}
	if _, err := svc.GetGitHost(user, connect.NewRequest(&agentsv1.GetGitHostRequest{Id: id})); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if _, err := svc.DeleteGitHost(admin, connect.NewRequest(&agentsv1.DeleteGitHostRequest{Id: id})); err != nil {
		t.Fatalf("admin Delete: %v", err)
	}
	_, err = svc.GetGitHost(user, connect.NewRequest(&agentsv1.GetGitHostRequest{Id: id}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestGitHostValidation(t *testing.T) {
	svc := newGitHostService()
	admin := auth.WithAdmin(context.Background())

	cases := []struct {
		name   string
		mutate func(*agentsv1.GitHost)
	}{
		{"missing name", func(h *agentsv1.GitHost) { h.Name = "" }},
		{"unspecified kind", func(h *agentsv1.GitHost) { h.Kind = agentsv1.GitHostKind_GIT_HOST_KIND_UNSPECIFIED }},
		{"relative api url", func(h *agentsv1.GitHost) { h.ApiBaseUrl = "api.github.com" }},
		{"non-http scheme", func(h *agentsv1.GitHost) { h.ApiBaseUrl = "ftp://api.github.com" }},
		{"bad web url", func(h *agentsv1.GitHost) { h.WebBaseUrl = "not a url" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := validHost()
			tc.mutate(h)
			_, err := svc.CreateGitHost(admin, connect.NewRequest(&agentsv1.CreateGitHostRequest{Host: h}))
			wantCode(t, err, connect.CodeInvalidArgument)
		})
	}
}
