package application

import (
	"context"
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/gitprovider"
	"go.orx.me/apps/butter/internal/repo/auth"
	githostrepo "go.orx.me/apps/butter/internal/repo/githost"
	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Validation check names reported by ValidateWorkspaceRepoBinding.
const (
	checkRepositoryRead      = "repository_read"
	checkBranchExists        = "branch_exists"
	checkWriteCapability     = "write_capability"
	checkChangeRequest       = "change_request_capability"
	repoContentSchemaV1      = 1
	notCheckedDetail         = "not checked: repository unreachable"
	credentialRejectedDetail = "credential rejected by host (expired or revoked?)"
)

// RepoBindingServiceServer implements
// agentsv1connect.WorkspaceRepoBindingServiceHandler (issue #214). Get is
// open to every workspace member (the auth middleware already verified
// membership for the X-Workspace-ID header); mutations require the caller to
// hold the workspace "owner" or "admin" role. PATs are encrypted with the
// configured key before storage and never appear in responses, logs, or
// error details.
type RepoBindingServiceServer struct {
	repo          repobindingrepo.Repository
	hostRepo      githostrepo.Repository
	workspaceRepo workspacerepo.Repository
	// encryptionKey returns the configured PAT encryption key. Lazy because
	// SetupRoutes runs before the YAML config is loaded.
	encryptionKey func() string
	// newProviderClient builds the git provider client; tests substitute a
	// fake. Defaults to gitprovider.New.
	newProviderClient func(gitprovider.Config) (gitprovider.Client, error)
}

func NewRepoBindingServiceServer(repo repobindingrepo.Repository, hostRepo githostrepo.Repository) *RepoBindingServiceServer {
	return &RepoBindingServiceServer{
		repo:              repo,
		hostRepo:          hostRepo,
		encryptionKey:     func() string { return "" },
		newProviderClient: gitprovider.New,
	}
}

// SetRepos wires the repositories after bootstrap.
func (s *RepoBindingServiceServer) SetRepos(repo repobindingrepo.Repository, hostRepo githostrepo.Repository) {
	s.repo = repo
	s.hostRepo = hostRepo
}

// SetWorkspaceRepo wires the workspace repository used for role checks.
func (s *RepoBindingServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

// SetEncryptionKeyProvider wires the lazy PAT encryption key source.
func (s *RepoBindingServiceServer) SetEncryptionKeyProvider(fn func() string) {
	if fn != nil {
		s.encryptionKey = fn
	}
}

// SetProviderClientFactory overrides the git provider client constructor
// (used by tests to substitute a fake provider).
func (s *RepoBindingServiceServer) SetProviderClientFactory(fn func(gitprovider.Config) (gitprovider.Client, error)) {
	if fn != nil {
		s.newProviderClient = fn
	}
}

func (s *RepoBindingServiceServer) requireRepos() error {
	if s.repo == nil || s.hostRepo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("repo binding repository not configured"))
	}
	return nil
}

func (s *RepoBindingServiceServer) cipher() (*secretbox.Cipher, error) {
	key := s.encryptionKey()
	if strings.TrimSpace(key) == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("git credential encryption key is not configured (set git.encryption_key)"))
	}
	c, err := secretbox.NewCipher(key)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("git credential encryption key is invalid"))
	}
	return c, nil
}

// requireBindingRole grants workspace "owner"/"admin" members and global
// admins (the bypass is audited). Members lacking the role receive
// PermissionDenied; non-members NotFound (middleware normally rejects them
// before this point).
func (s *RepoBindingServiceServer) requireBindingRole(ctx context.Context, workspaceID string) error {
	if auth.IsAdmin(ctx) {
		if user, ok := auth.UserFromContext(ctx); ok {
			log.FromContext(ctx).Info("global admin managing workspace repo binding",
				"audit", "admin_repo_binding_access", "workspace_id", workspaceID, "user_id", user.GetId())
		}
		return nil
	}
	if s.workspaceRepo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace repository not configured"))
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, user.GetId())
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

func mapRepoBindingErr(err error) *connect.Error {
	if errors.Is(err, repobindingrepo.ErrNotFound) {
		return connectx.NotFound("workspace has no repository binding")
	}
	return connectx.InternalWith(err)
}

// normalizeRootPath cleans and validates the repository-relative root
// directory. Empty means the repository root.
func normalizeRootPath(raw string) (string, error) {
	p := strings.TrimSpace(raw)
	if p == "" {
		return "", nil
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return "", connectx.InvalidArgument("root_path", "must be a repository-relative directory")
	}
	clean := path.Clean(p)
	if clean == "." {
		return "", nil
	}
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", connectx.InvalidArgument("root_path", "must not traverse outside the repository")
	}
	return clean, nil
}

func providerKind(kind agentsv1.GitHostKind) (gitprovider.Kind, error) {
	switch kind {
	case agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB:
		return gitprovider.KindGitHub, nil
	case agentsv1.GitHostKind_GIT_HOST_KIND_GITLAB:
		return gitprovider.KindGitLab, nil
	default:
		return "", errors.New("git host has no supported provider kind")
	}
}

// sanitizeBinding validates and normalizes caller-supplied binding fields,
// dropping every server-owned field.
func (s *RepoBindingServiceServer) sanitizeBinding(ctx context.Context, in *agentsv1.WorkspaceRepoBinding) (*agentsv1.WorkspaceRepoBinding, error) {
	if in == nil {
		return nil, connectx.RequiredArgument("binding")
	}
	if strings.TrimSpace(in.GetGitHostId()) == "" {
		return nil, connectx.RequiredArgument("binding.git_host_id")
	}
	repository := strings.Trim(strings.TrimSpace(in.GetRepository()), "/")
	if repository == "" {
		return nil, connectx.RequiredArgument("binding.repository")
	}
	if !strings.Contains(repository, "/") {
		return nil, connectx.InvalidArgument("binding.repository", "must include its namespace (owner/repo)")
	}
	branch := strings.TrimSpace(in.GetBranch())
	if branch == "" {
		return nil, connectx.RequiredArgument("binding.branch")
	}
	rootPath, err := normalizeRootPath(in.GetRootPath())
	if err != nil {
		return nil, err
	}
	writeMode := in.GetWriteMode()
	if writeMode == agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_UNSPECIFIED {
		writeMode = agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_DIRECT_COMMIT
	}
	schema := in.GetContentSchemaVersion()
	if schema == 0 {
		schema = repoContentSchemaV1
	}
	if schema != repoContentSchemaV1 {
		return nil, connectx.InvalidArgument("binding.content_schema_version", "only version 1 is supported")
	}
	if _, err := s.hostRepo.Get(ctx, in.GetGitHostId()); err != nil {
		if errors.Is(err, githostrepo.ErrNotFound) {
			return nil, connectx.InvalidArgument("binding.git_host_id", "unknown git host")
		}
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.WorkspaceRepoBinding{
		GitHostId:            in.GetGitHostId(),
		Repository:           repository,
		Branch:               branch,
		RootPath:             rootPath,
		WriteMode:            writeMode,
		ContentSchemaVersion: schema,
		Status: &agentsv1.RepoBindingStatus{
			State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_UNVALIDATED,
		},
	}, nil
}

func (s *RepoBindingServiceServer) GetWorkspaceRepoBinding(ctx context.Context, _ *connect.Request[agentsv1.GetWorkspaceRepoBindingRequest]) (*connect.Response[agentsv1.GetWorkspaceRepoBindingResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	binding, err := s.repo.Get(ctx, ws)
	if errors.Is(err, repobindingrepo.ErrNotFound) {
		return connect.NewResponse(&agentsv1.GetWorkspaceRepoBindingResponse{}), nil
	}
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	overlaps, err := s.findOverlaps(ctx, binding)
	if err != nil {
		log.FromContext(ctx).Error("compute repo binding overlaps failed", "workspace_id", ws, "err", err)
		overlaps = nil
	}
	return connect.NewResponse(&agentsv1.GetWorkspaceRepoBindingResponse{
		Binding:  binding,
		Overlaps: overlaps,
	}), nil
}

// findOverlaps lists bindings in other workspaces that resolve to the same
// effective repository location. Overlap is allowed and intentional (shared
// Agent Content); it is surfaced so it is never a surprise.
func (s *RepoBindingServiceServer) findOverlaps(ctx context.Context, binding *agentsv1.WorkspaceRepoBinding) ([]*agentsv1.RepoBindingOverlap, error) {
	all, err := s.repo.ListAcrossWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	var out []*agentsv1.RepoBindingOverlap
	for _, other := range all {
		if other.GetWorkspaceId() == binding.GetWorkspaceId() {
			continue
		}
		if other.GetGitHostId() != binding.GetGitHostId() ||
			other.GetRepository() != binding.GetRepository() ||
			other.GetBranch() != binding.GetBranch() ||
			other.GetRootPath() != binding.GetRootPath() {
			continue
		}
		overlap := &agentsv1.RepoBindingOverlap{WorkspaceId: other.GetWorkspaceId()}
		if s.workspaceRepo != nil {
			if w, err := s.workspaceRepo.GetWorkspace(ctx, other.GetWorkspaceId()); err == nil {
				overlap.WorkspaceName = w.GetName()
			}
		}
		out = append(out, overlap)
	}
	return out, nil
}

func (s *RepoBindingServiceServer) PutWorkspaceRepoBinding(ctx context.Context, req *connect.Request[agentsv1.PutWorkspaceRepoBindingRequest]) (*connect.Response[agentsv1.PutWorkspaceRepoBindingResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireBindingRole(ctx, ws); err != nil {
		return nil, err
	}
	binding, err := s.sanitizeBinding(ctx, req.Msg.GetBinding())
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	stored, err := s.repo.Put(ctx, ws, binding)
	if err != nil {
		logger.Error("put repo binding failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	logger.Info("repo binding saved", "workspace_id", ws,
		"git_host_id", stored.GetGitHostId(), "repository", stored.GetRepository(),
		"branch", stored.GetBranch(), "root_path", stored.GetRootPath(),
		"write_mode", stored.GetWriteMode().String())
	return connect.NewResponse(&agentsv1.PutWorkspaceRepoBindingResponse{Binding: stored}), nil
}

func (s *RepoBindingServiceServer) DeleteWorkspaceRepoBinding(ctx context.Context, _ *connect.Request[agentsv1.DeleteWorkspaceRepoBindingRequest]) (*connect.Response[agentsv1.DeleteWorkspaceRepoBindingResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireBindingRole(ctx, ws); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	if err := s.repo.Delete(ctx, ws); err != nil {
		logger.Error("delete repo binding failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	logger.Info("repo binding deleted", "workspace_id", ws)
	return connect.NewResponse(&agentsv1.DeleteWorkspaceRepoBindingResponse{}), nil
}

func (s *RepoBindingServiceServer) SetWorkspaceRepoBindingCredential(ctx context.Context, req *connect.Request[agentsv1.SetWorkspaceRepoBindingCredentialRequest]) (*connect.Response[agentsv1.SetWorkspaceRepoBindingCredentialResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireBindingRole(ctx, ws); err != nil {
		return nil, err
	}
	pat := strings.TrimSpace(req.Msg.GetPat())
	if pat == "" {
		return nil, connectx.RequiredArgument("pat")
	}
	cipher, err := s.cipher()
	if err != nil {
		return nil, err
	}
	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}
	ciphertext, encErr := cipher.Encrypt([]byte(pat))
	if encErr != nil {
		return nil, connectx.Internal("failed to encrypt credential")
	}
	logger := log.FromContext(ctx)
	if err := s.repo.SetCredential(ctx, ws, ciphertext); err != nil {
		logger.Error("set repo binding credential failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	// A new credential invalidates any previous validation result.
	binding.Status = &agentsv1.RepoBindingStatus{
		State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_UNVALIDATED,
	}
	stored, err := s.repo.Put(ctx, ws, binding)
	if err != nil {
		logger.Error("reset repo binding status failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	logger.Info("repo binding credential set", "workspace_id", ws)
	return connect.NewResponse(&agentsv1.SetWorkspaceRepoBindingCredentialResponse{Binding: stored}), nil
}

func (s *RepoBindingServiceServer) ValidateWorkspaceRepoBinding(ctx context.Context, _ *connect.Request[agentsv1.ValidateWorkspaceRepoBindingRequest]) (*connect.Response[agentsv1.ValidateWorkspaceRepoBindingResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireBindingRole(ctx, ws); err != nil {
		return nil, err
	}
	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}
	host, err := s.hostRepo.Get(ctx, binding.GetGitHostId())
	if err != nil {
		if errors.Is(err, githostrepo.ErrNotFound) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("the binding's git host is no longer configured"))
		}
		return nil, connectx.InternalWith(err)
	}
	ciphertext, err := s.repo.GetCredential(ctx, ws)
	if err != nil {
		if errors.Is(err, repobindingrepo.ErrNoCredential) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("set a credential before validating the binding"))
		}
		return nil, mapRepoBindingErr(err)
	}
	cipher, err := s.cipher()
	if err != nil {
		return nil, err
	}
	pat, decErr := cipher.Decrypt(ciphertext)
	if decErr != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("stored credential cannot be decrypted (encryption key changed?); replace the credential"))
	}
	kind, kindErr := providerKind(host.GetKind())
	if kindErr != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, kindErr)
	}
	client, err := s.newProviderClient(gitprovider.Config{
		Kind:       kind,
		APIBaseURL: host.GetApiBaseUrl(),
		Repository: binding.GetRepository(),
		Token:      string(pat),
	})
	if err != nil {
		return nil, connectx.InvalidArgument("binding", "binding cannot be validated: "+err.Error())
	}

	status := runBindingChecks(ctx, client, binding)
	binding.Status = status
	logger := log.FromContext(ctx)
	stored, err := s.repo.Put(ctx, ws, binding)
	if err != nil {
		logger.Error("persist repo binding validation failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	logger.Info("repo binding validated", "workspace_id", ws, "state", status.GetState().String(), "error", status.GetError())
	return connect.NewResponse(&agentsv1.ValidateWorkspaceRepoBindingResponse{Binding: stored}), nil
}

// providerErrDetail maps a provider error onto a stable, credential-free
// human-readable detail string.
func providerErrDetail(err error) string {
	switch {
	case errors.Is(err, gitprovider.ErrUnauthorized):
		return credentialRejectedDetail
	case errors.Is(err, gitprovider.ErrForbidden):
		return "access forbidden for this credential"
	case errors.Is(err, gitprovider.ErrNotFound):
		return "not found or not visible to this credential"
	default:
		return "provider request failed"
	}
}

// runBindingChecks probes the repository through the provider seam and
// composes the persisted status. Required checks depend on the write mode:
// read and branch always gate; write capability always gates (both modes
// push commits); change-request capability gates only in CHANGE_REQUEST
// mode but is always reported.
func runBindingChecks(ctx context.Context, client gitprovider.Client, binding *agentsv1.WorkspaceRepoBinding) *agentsv1.RepoBindingStatus {
	changeRequestRequired := binding.GetWriteMode() == agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_CHANGE_REQUEST

	readCheck := &agentsv1.RepoBindingCheck{Name: checkRepositoryRead, Required: true}
	branchCheck := &agentsv1.RepoBindingCheck{Name: checkBranchExists, Required: true}
	writeCheck := &agentsv1.RepoBindingCheck{Name: checkWriteCapability, Required: true}
	crCheck := &agentsv1.RepoBindingCheck{Name: checkChangeRequest, Required: changeRequestRequired}

	repoInfo, err := client.GetRepository(ctx)
	if err != nil {
		readCheck.Detail = providerErrDetail(err)
		branchCheck.Detail = notCheckedDetail
		writeCheck.Detail = notCheckedDetail
		crCheck.Detail = notCheckedDetail
	} else {
		readCheck.Ok = true
		visibility := "private"
		if !repoInfo.Private {
			visibility = "public"
		}
		readCheck.Detail = "repository is " + visibility

		if _, err := client.GetBranchHead(ctx, binding.GetBranch()); err != nil {
			branchCheck.Detail = fmt.Sprintf("branch %q %s", binding.GetBranch(), providerErrDetail(err))
		} else {
			branchCheck.Ok = true
			branchCheck.Detail = fmt.Sprintf("branch %q exists", binding.GetBranch())
		}

		writeCheck.Ok = repoInfo.CanWrite
		if !repoInfo.CanWrite {
			writeCheck.Detail = "credential has no push access"
		}
		crCheck.Ok = repoInfo.CanOpenChangeRequests
		if !repoInfo.CanOpenChangeRequests {
			crCheck.Detail = "credential cannot open pull/merge requests"
		}
	}

	checks := []*agentsv1.RepoBindingCheck{readCheck, branchCheck, writeCheck, crCheck}
	status := &agentsv1.RepoBindingStatus{
		State:           agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_OK,
		LastValidatedAt: timestamppb.New(time.Now().UTC()),
		Checks:          checks,
	}
	for _, c := range checks {
		if c.GetRequired() && !c.GetOk() {
			status.State = agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_FAILED
			if status.Error == "" {
				status.Error = c.GetName() + ": " + c.GetDetail()
			}
		}
	}
	return status
}
