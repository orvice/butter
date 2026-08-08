package application

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/repo/auth"
	githostrepo "go.orx.me/apps/butter/internal/repo/githost"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// GitHostServiceServer implements agentsv1connect.GitHostServiceHandler.
// Hosts form the platform-level allowlist of Git endpoints workspaces may
// bind repositories from (issue #214): List/Get serve any authenticated
// user (owners need them to configure a binding), mutations require global
// admin so workspace input can never introduce an arbitrary API base URL.
type GitHostServiceServer struct {
	repo githostrepo.Repository
}

func NewGitHostServiceServer(repo githostrepo.Repository) *GitHostServiceServer {
	return &GitHostServiceServer{repo: repo}
}

// SetRepo wires the repository after bootstrap.
func (s *GitHostServiceServer) SetRepo(repo githostrepo.Repository) {
	s.repo = repo
}

func (s *GitHostServiceServer) requireRepo() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("git host repository not configured"))
	}
	return nil
}

func requireGlobalAdmin(ctx context.Context) error {
	if !auth.IsAdmin(ctx) {
		return connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	return nil
}

func mapGitHostErr(err error) *connect.Error {
	if errors.Is(err, githostrepo.ErrNotFound) {
		return connectx.NotFound(err.Error())
	}
	if errors.Is(err, githostrepo.ErrAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}
	return connectx.InternalWith(err)
}

// validateGitHost checks admin-supplied host fields. The API base URL must
// be an absolute http(s) URL; anything else is rejected before storage.
func validateGitHost(h *agentsv1.GitHost) error {
	if h == nil {
		return connectx.RequiredArgument("host")
	}
	if strings.TrimSpace(h.GetName()) == "" {
		return connectx.RequiredArgument("name")
	}
	switch h.GetKind() {
	case agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB, agentsv1.GitHostKind_GIT_HOST_KIND_GITLAB:
	default:
		return connectx.InvalidArgument("kind", "must be GITHUB or GITLAB")
	}
	u, err := url.Parse(strings.TrimSpace(h.GetApiBaseUrl()))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return connectx.InvalidArgument("api_base_url", "must be an absolute http(s) URL")
	}
	if web := strings.TrimSpace(h.GetWebBaseUrl()); web != "" {
		w, err := url.Parse(web)
		if err != nil || (w.Scheme != "http" && w.Scheme != "https") || w.Host == "" {
			return connectx.InvalidArgument("web_base_url", "must be an absolute http(s) URL")
		}
	}
	return nil
}

func (s *GitHostServiceServer) ListGitHosts(ctx context.Context, _ *connect.Request[agentsv1.ListGitHostsRequest]) (*connect.Response[agentsv1.ListGitHostsResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	hosts, err := s.repo.List(ctx)
	if err != nil {
		return nil, mapGitHostErr(err)
	}
	return connect.NewResponse(&agentsv1.ListGitHostsResponse{Hosts: hosts}), nil
}

func (s *GitHostServiceServer) GetGitHost(ctx context.Context, req *connect.Request[agentsv1.GetGitHostRequest]) (*connect.Response[agentsv1.GetGitHostResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	host, err := s.repo.Get(ctx, req.Msg.GetId())
	if err != nil {
		return nil, mapGitHostErr(err)
	}
	return connect.NewResponse(&agentsv1.GetGitHostResponse{Host: host}), nil
}

func (s *GitHostServiceServer) CreateGitHost(ctx context.Context, req *connect.Request[agentsv1.CreateGitHostRequest]) (*connect.Response[agentsv1.CreateGitHostResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}
	host := req.Msg.GetHost()
	if err := validateGitHost(host); err != nil {
		return nil, err
	}
	if host.GetId() == "" {
		host.Id = uuid.NewString()
	}
	logger := log.FromContext(ctx)
	created, err := s.repo.Create(ctx, host)
	if err != nil {
		logger.Error("create git host failed", "id", host.GetId(), "err", err)
		return nil, mapGitHostErr(err)
	}
	logger.Info("git host created", "id", created.GetId(), "name", created.GetName(), "kind", created.GetKind().String())
	return connect.NewResponse(&agentsv1.CreateGitHostResponse{Host: created}), nil
}

func (s *GitHostServiceServer) UpdateGitHost(ctx context.Context, req *connect.Request[agentsv1.UpdateGitHostRequest]) (*connect.Response[agentsv1.UpdateGitHostResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}
	host := req.Msg.GetHost()
	if host.GetId() == "" {
		return nil, connectx.RequiredArgument("host.id")
	}
	if err := validateGitHost(host); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	updated, err := s.repo.Update(ctx, host)
	if err != nil {
		logger.Error("update git host failed", "id", host.GetId(), "err", err)
		return nil, mapGitHostErr(err)
	}
	logger.Info("git host updated", "id", updated.GetId(), "name", updated.GetName())
	return connect.NewResponse(&agentsv1.UpdateGitHostResponse{Host: updated}), nil
}

func (s *GitHostServiceServer) DeleteGitHost(ctx context.Context, req *connect.Request[agentsv1.DeleteGitHostRequest]) (*connect.Response[agentsv1.DeleteGitHostResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	logger := log.FromContext(ctx)
	if err := s.repo.Delete(ctx, req.Msg.GetId()); err != nil {
		logger.Error("delete git host failed", "id", req.Msg.GetId(), "err", err)
		return nil, mapGitHostErr(err)
	}
	logger.Info("git host deleted", "id", req.Msg.GetId())
	return connect.NewResponse(&agentsv1.DeleteGitHostResponse{}), nil
}
