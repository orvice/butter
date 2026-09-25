package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/mem0"
	"go.orx.me/apps/butter/internal/repo/auth"
	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// workspaceMemoryProbeTimeout bounds one connection probe. A search is one
// embedding call on the mem0 side, not an extraction.
const workspaceMemoryProbeTimeout = 10 * time.Second

// workspaceMemoryProbeUserID scopes the probe search to an identity no
// Workspace Memory is ever written under, so the probe reads nothing.
const workspaceMemoryProbeUserID = "butter:probe"

// WorkspaceMemoryConfigServiceServer implements
// agentsv1connect.WorkspaceMemoryConfigServiceHandler (ADR-0013). The config
// is a per-workspace singleton; its mem0 API key lives behind the secretbox
// credential seam and is never read back.
type WorkspaceMemoryConfigServiceServer struct {
	repo          memoryconfigrepo.Repository
	keyring       *secretbox.Keyring
	workspaceRepo workspacerepo.Repository
	// httpClient carries probe requests; tests point it at a fake server.
	httpClient *http.Client
}

func NewWorkspaceMemoryConfigServiceServer(repo memoryconfigrepo.Repository) *WorkspaceMemoryConfigServiceServer {
	return &WorkspaceMemoryConfigServiceServer{repo: repo, httpClient: &http.Client{}}
}

// SetRepo wires the repository after bootstrap.
func (s *WorkspaceMemoryConfigServiceServer) SetRepo(repo memoryconfigrepo.Repository) { s.repo = repo }

// SetKeyring wires credential encryption after bootstrap. Without it, key
// writes are refused (a nil keyring fails closed).
func (s *WorkspaceMemoryConfigServiceServer) SetKeyring(k *secretbox.Keyring) { s.keyring = k }

// SetWorkspaceRepo wires the membership lookup behind the manage-role check.
func (s *WorkspaceMemoryConfigServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

func (s *WorkspaceMemoryConfigServiceServer) requireRepo() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace memory config repository not configured"))
	}
	return nil
}

func mapWorkspaceMemoryErr(err error) *connect.Error {
	if errors.Is(err, memoryconfigrepo.ErrNotFound) {
		return connectx.NotFound("workspace has no memory config")
	}
	return connectx.InternalWith(err)
}

// validateMemoryBaseURL requires an absolute http(s) URL.
func validateMemoryBaseURL(baseURL string) error {
	if strings.TrimSpace(baseURL) == "" {
		return connectx.RequiredArgument("base_url")
	}
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return connectx.InvalidArgument("base_url", "must be an absolute http(s) URL")
	}
	return nil
}

// probe runs one read-only search against the server. A nil error means the
// server answered and accepted the key.
func (s *WorkspaceMemoryConfigServiceServer) probe(ctx context.Context, baseURL, apiKey string) error {
	probeCtx, cancel := context.WithTimeout(ctx, workspaceMemoryProbeTimeout)
	defer cancel()
	_, err := mem0.New(baseURL, apiKey, s.httpClient).Search(probeCtx, mem0.SearchRequest{
		Query:   "butter connection probe",
		Filters: map[string]any{"user_id": workspaceMemoryProbeUserID},
		TopK:    1,
	})
	return err
}

func (s *WorkspaceMemoryConfigServiceServer) GetWorkspaceMemoryConfig(ctx context.Context, _ *connect.Request[agentsv1.GetWorkspaceMemoryConfigRequest]) (*connect.Response[agentsv1.GetWorkspaceMemoryConfigResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	cfg, err := s.repo.Get(ctx, workspaceID)
	if errors.Is(err, memoryconfigrepo.ErrNotFound) {
		return connect.NewResponse(&agentsv1.GetWorkspaceMemoryConfigResponse{}), nil
	}
	if err != nil {
		return nil, mapWorkspaceMemoryErr(err)
	}
	return connect.NewResponse(&agentsv1.GetWorkspaceMemoryConfigResponse{Config: cfg}), nil
}

func (s *WorkspaceMemoryConfigServiceServer) PutWorkspaceMemoryConfig(ctx context.Context, req *connect.Request[agentsv1.PutWorkspaceMemoryConfigRequest]) (*connect.Response[agentsv1.PutWorkspaceMemoryConfigResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "workspace_memory"); err != nil {
		return nil, err
	}
	if err := validateMemoryBaseURL(req.Msg.GetBaseUrl()); err != nil {
		return nil, err
	}
	baseURL := normalizeBaseURL(req.Msg.GetBaseUrl())

	// The probe must use the key the config will hold after this call: the
	// new one when supplied, otherwise the stored one.
	var probeKey string
	if req.Msg.ApiKey != nil {
		probeKey = req.Msg.GetApiKey()
	} else if req.Msg.GetEnabled() {
		conn, err := memoryconn.NewResolver(s.repo, s.keyring).Load(ctx, workspaceID)
		if err != nil && !errors.Is(err, memoryconn.ErrNotConfigured) {
			return nil, connectx.InternalWith(err)
		}
		probeKey = conn.APIKey
	}

	warning := ""
	if req.Msg.GetEnabled() {
		if err := s.probe(ctx, baseURL, probeKey); err != nil {
			if mem0.IsAuthError(err) {
				return nil, connect.NewError(connect.CodeFailedPrecondition,
					fmt.Errorf("mem0 server rejected the API key; check the key and that it is valid on %s: %w", baseURL, err))
			}
			warning = fmt.Sprintf("saved, but the mem0 server did not answer the connection probe: %v", err)
		}
	}

	var cred *memoryconfigrepo.Credential
	if req.Msg.ApiKey != nil {
		cred = &memoryconfigrepo.Credential{}
		if key := req.Msg.GetApiKey(); key != "" {
			if s.keyring == nil {
				return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("credential encryption is not configured"))
			}
			ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(key))
			if err != nil {
				return nil, connectx.InternalWith(fmt.Errorf("encrypt mem0 API key: %w", err))
			}
			cred = &memoryconfigrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}
		}
	}

	saved, err := s.repo.Put(ctx, workspaceID, &agentsv1.WorkspaceMemoryConfig{
		BaseUrl: baseURL,
		Enabled: req.Msg.GetEnabled(),
	}, cred)
	if err != nil {
		return nil, mapWorkspaceMemoryErr(err)
	}
	log.FromContext(ctx).Info("workspace memory config saved",
		"workspace", workspaceID, "enabled", saved.GetEnabled(),
		"key_changed", cred != nil, "probe_warning", warning != "")
	return connect.NewResponse(&agentsv1.PutWorkspaceMemoryConfigResponse{Config: saved, Warning: warning}), nil
}

func (s *WorkspaceMemoryConfigServiceServer) DeleteWorkspaceMemoryConfig(ctx context.Context, _ *connect.Request[agentsv1.DeleteWorkspaceMemoryConfigRequest]) (*connect.Response[agentsv1.DeleteWorkspaceMemoryConfigResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "workspace_memory"); err != nil {
		return nil, err
	}
	// No reference check: memory-enabled agents degrade to running without
	// memory (ADR-0013 §2).
	if err := s.repo.Delete(ctx, workspaceID); err != nil {
		return nil, mapWorkspaceMemoryErr(err)
	}
	log.FromContext(ctx).Info("workspace memory config deleted", "workspace", workspaceID)
	return connect.NewResponse(&agentsv1.DeleteWorkspaceMemoryConfigResponse{}), nil
}

func (s *WorkspaceMemoryConfigServiceServer) TestWorkspaceMemoryConnection(ctx context.Context, _ *connect.Request[agentsv1.TestWorkspaceMemoryConnectionRequest]) (*connect.Response[agentsv1.TestWorkspaceMemoryConnectionResponse], error) {
	if err := s.requireRepo(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := memoryconn.NewResolver(s.repo, s.keyring).Load(ctx, workspaceID)
	if errors.Is(err, memoryconn.ErrNotConfigured) {
		return nil, mapWorkspaceMemoryErr(memoryconfigrepo.ErrNotFound)
	}
	if err != nil {
		// An undecryptable key is a probe outcome the settings view must
		// render, not a failure to report it.
		return connect.NewResponse(&agentsv1.TestWorkspaceMemoryConnectionResponse{Error: err.Error()}), nil
	}
	if err := s.probe(ctx, conn.BaseURL, conn.APIKey); err != nil {
		return connect.NewResponse(&agentsv1.TestWorkspaceMemoryConnectionResponse{Error: err.Error()}), nil
	}
	return connect.NewResponse(&agentsv1.TestWorkspaceMemoryConnectionResponse{Ok: true}), nil
}

// requireWorkspaceManageRole grants workspace "owner"/"admin" members and
// global admins (the bypass is audited under `admin_<resource>_access`).
// Members lacking the role receive PermissionDenied; non-members NotFound,
// which is what the middleware would already have returned for a workspace
// they cannot see.
func requireWorkspaceManageRole(ctx context.Context, wsRepo workspacerepo.Repository, workspaceID, resource string) error {
	if auth.IsAdmin(ctx) {
		if user, ok := auth.UserFromContext(ctx); ok {
			log.FromContext(ctx).Info("global admin managing workspace resource",
				"audit", "admin_"+resource+"_access", "resource", resource,
				"workspace_id", workspaceID, "user_id", user.GetId())
		}
		return nil
	}
	if wsRepo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace repository not configured"))
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	member, err := wsRepo.GetMember(ctx, workspaceID, user.GetId())
	if err != nil {
		if errors.Is(err, workspacerepo.ErrNotFound) {
			return connectx.NotFound("workspace not found")
		}
		return connectx.InternalWith(err)
	}
	if slices.Contains([]string{"owner", "admin"}, member.GetRole()) {
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("workspace owner or admin role required"))
}
