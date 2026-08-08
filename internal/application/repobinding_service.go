package application

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/agentcontent"
	"go.orx.me/apps/butter/internal/gitprovider"
	"go.orx.me/apps/butter/internal/repo/auth"
	agentcontentrepo "go.orx.me/apps/butter/internal/repo/agentcontent"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	githostrepo "go.orx.me/apps/butter/internal/repo/githost"
	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	"go.orx.me/apps/butter/internal/repo/repocache"
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

	defaultMaxFileBytes      = 256 * 1024       // 256 KiB per file
	defaultMaxWorkspaceCache = 20 * 1024 * 1024 // 20 MiB per workspace
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
	agentRepo     configrepo.AgentRepository
	cacheRepo     repocache.Repository
	contentRepo   agentcontentrepo.Repository
	// encryptionKey returns the configured PAT encryption key. Lazy because
	// SetupRoutes runs before the YAML config is loaded.
	encryptionKey func() string
	// newProviderClient builds the git provider client; tests substitute a
	// fake. Defaults to gitprovider.New.
	newProviderClient func(gitprovider.Config) (gitprovider.Client, error)
	cacheLimits       func() (maxFileBytes, maxCacheBytes int64)
	// configRuntime triggers runner reload after successful publication.
	configRuntime ConfigRuntime
	// webhookBaseURL is the externally reachable base URL for webhook
	// callbacks, e.g. "https://butter.example.com".
	webhookBaseURL func() string
	// publishMu serializes publication per workspace (single-instance; a
	// distributed lease would replace this for multi-instance deployments).
	publishMu sync.Map // workspace_id → *sync.Mutex
}

func NewRepoBindingServiceServer(repo repobindingrepo.Repository, hostRepo githostrepo.Repository) *RepoBindingServiceServer {
	return &RepoBindingServiceServer{
		repo:              repo,
		hostRepo:          hostRepo,
		encryptionKey:     func() string { return "" },
		newProviderClient: gitprovider.New,
		cacheLimits: func() (int64, int64) {
			return defaultMaxFileBytes, defaultMaxWorkspaceCache
		},
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

// SetAgentRepo wires the workspace-scoped Agent repository used to mark
// cached agents/<agent-id> paths as claimed.
func (s *RepoBindingServiceServer) SetAgentRepo(repo configrepo.AgentRepository) {
	s.agentRepo = repo
}

// SetEncryptionKeyProvider wires the lazy PAT encryption key source.
func (s *RepoBindingServiceServer) SetEncryptionKeyProvider(fn func() string) {
	if fn != nil {
		s.encryptionKey = fn
	}
}

// SetCacheRepo wires the repository cache storage.
func (s *RepoBindingServiceServer) SetCacheRepo(repo repocache.Repository) {
	s.cacheRepo = repo
}

// SetCacheLimitsProvider wires lazily loaded repository cache limits.
func (s *RepoBindingServiceServer) SetCacheLimitsProvider(fn func() (int64, int64)) {
	if fn != nil {
		s.cacheLimits = fn
	}
}

// SetProviderClientFactory overrides the git provider client constructor
// (used by tests to substitute a fake provider).
func (s *RepoBindingServiceServer) SetProviderClientFactory(fn func(gitprovider.Config) (gitprovider.Client, error)) {
	if fn != nil {
		s.newProviderClient = fn
	}
}

// SetContentRepo wires the agent content snapshot storage.
func (s *RepoBindingServiceServer) SetContentRepo(repo agentcontentrepo.Repository) {
	s.contentRepo = repo
}

// SetConfigRuntime wires the runtime reload trigger used after publication.
func (s *RepoBindingServiceServer) SetConfigRuntime(rt ConfigRuntime) {
	s.configRuntime = rt
}

// SetWebhookBaseURL wires the lazy webhook callback base URL provider.
func (s *RepoBindingServiceServer) SetWebhookBaseURL(fn func() string) {
	if fn != nil {
		s.webhookBaseURL = fn
	}
}

// publishMutex returns or creates the per-workspace publication mutex.
func (s *RepoBindingServiceServer) publishMutex(workspaceID string) *sync.Mutex {
	v, _ := s.publishMu.LoadOrStore(workspaceID, &sync.Mutex{})
	return v.(*sync.Mutex)
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

// normalizeRootPath validates the repository-relative root directory. Empty
// means the repository root. Traversal is rejected outright rather than
// cleaned (issue #210: repository paths reject absolute paths and
// traversal), so "a/../b" is an error, not "b".
func normalizeRootPath(raw string) (string, error) {
	p := strings.TrimSpace(raw)
	if p == "" {
		return "", nil
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return "", connectx.InvalidArgument("root_path", "must be a repository-relative directory")
	}
	if slices.Contains(strings.Split(p, "/"), "..") {
		return "", connectx.InvalidArgument("root_path", "must not contain traversal segments")
	}
	clean := path.Clean(p)
	if clean == "." {
		return "", nil
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

func repoCacheBindingKey(binding *agentsv1.WorkspaceRepoBinding) string {
	location := strings.Join([]string{
		binding.GetGitHostId(),
		strings.ToLower(binding.GetRepository()),
		binding.GetBranch(),
		binding.GetRootPath(),
		fmt.Sprint(binding.GetContentSchemaVersion()),
	}, "\x00")
	sum := sha256.Sum256([]byte(location))
	return hex.EncodeToString(sum[:])
}

func isClaimedAgentPath(entryPath string, knownAgentIDs map[string]struct{}) bool {
	parts := strings.Split(entryPath, "/")
	if len(parts) < 2 || parts[0] != "agents" {
		return false
	}
	_, ok := knownAgentIDs[parts[1]]
	return ok
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
	if strings.Contains(repository, "\\") || slices.Contains(strings.Split(repository, "/"), "..") {
		return nil, connectx.InvalidArgument("binding.repository", "must be a plain owner/repo path")
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
// Agent Content); it is surfaced so it is never a surprise. Hosts are
// compared by their API base URL (duplicate host records for the same
// endpoint still overlap) and repositories case-insensitively (GitHub and
// GitLab paths are case-insensitive); branches and root paths are
// case-sensitive like git itself.
func (s *RepoBindingServiceServer) findOverlaps(ctx context.Context, binding *agentsv1.WorkspaceRepoBinding) ([]*agentsv1.RepoBindingOverlap, error) {
	all, err := s.repo.ListAcrossWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	hostKeys := map[string]string{}
	if hosts, err := s.hostRepo.List(ctx); err == nil {
		for _, h := range hosts {
			hostKeys[h.GetId()] = strings.TrimRight(strings.ToLower(strings.TrimSpace(h.GetApiBaseUrl())), "/")
		}
	}
	locationKey := func(b *agentsv1.WorkspaceRepoBinding) string {
		hostKey, ok := hostKeys[b.GetGitHostId()]
		if !ok || hostKey == "" {
			hostKey = "id:" + b.GetGitHostId()
		}
		return strings.Join([]string{hostKey, strings.ToLower(b.GetRepository()), b.GetBranch(), b.GetRootPath()}, "\n")
	}
	key := locationKey(binding)
	var out []*agentsv1.RepoBindingOverlap
	for _, other := range all {
		if other.GetWorkspaceId() == binding.GetWorkspaceId() {
			continue
		}
		if locationKey(other) != key {
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
	// A PAT is scoped to one host: when the binding moves to a different
	// host, the stored credential must never be forwarded there. Clear it
	// before the Put so a partial failure leaves the old binding without a
	// misdirected credential rather than the new binding with one.
	if prev, err := s.repo.Get(ctx, ws); err == nil && prev.GetGitHostId() != binding.GetGitHostId() {
		if err := s.repo.SetCredential(ctx, ws, ""); err != nil {
			logger.Error("clear repo binding credential failed", "workspace_id", ws, "err", err)
			return nil, mapRepoBindingErr(err)
		}
		logger.Info("repo binding credential cleared (git host changed)", "workspace_id", ws)
	} else if err != nil && !errors.Is(err, repobindingrepo.ErrNotFound) {
		return nil, connectx.InternalWith(err)
	}
	stored, err := s.repo.Put(ctx, ws, binding)
	if err != nil {
		logger.Error("put repo binding failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	// Invalidate the repo cache whenever the binding changes; the old
	// cache may belong to a different repository/branch/root.
	if s.cacheRepo != nil {
		if delErr := s.cacheRepo.Delete(ctx, ws); delErr != nil {
			logger.Warn("failed to invalidate repo cache on binding update", "workspace_id", ws, "err", delErr)
		}
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
	// Purge cached content so stale data from the deleted binding is not
	// accessible.
	if s.cacheRepo != nil {
		if delErr := s.cacheRepo.Delete(ctx, ws); delErr != nil {
			logger.Warn("failed to invalidate repo cache on binding delete", "workspace_id", ws, "err", delErr)
		}
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
	// Reset validation status BEFORE storing the new credential (ADR-0005:
	// replacement resets to UNVALIDATED). If the credential write then
	// fails, the binding is left unvalidated with the old credential —
	// never a new credential wearing a stale OK status.
	binding.Status = &agentsv1.RepoBindingStatus{
		State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_UNVALIDATED,
	}
	if _, err := s.repo.Put(ctx, ws, binding); err != nil {
		logger.Error("reset repo binding status failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	if err := s.repo.SetCredential(ctx, ws, ciphertext); err != nil {
		logger.Error("set repo binding credential failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}
	stored, err := s.repo.Get(ctx, ws)
	if err != nil {
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

// resolveProviderClient builds a provider client from the binding's stored
// credential and associated git host. Shared by Validate and Sync.
func (s *RepoBindingServiceServer) resolveProviderClient(ctx context.Context, ws string, binding *agentsv1.WorkspaceRepoBinding) (gitprovider.Client, error) {
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
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("set a credential before syncing the repository"))
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
		return nil, connectx.InvalidArgument("binding", "cannot build provider client: "+err.Error())
	}
	return client, nil
}

// ── Sync and cache RPCs (issue #215) ────────────────────────────────────

func (s *RepoBindingServiceServer) SyncWorkspaceRepository(ctx context.Context, _ *connect.Request[agentsv1.SyncWorkspaceRepositoryRequest]) (*connect.Response[agentsv1.SyncWorkspaceRepositoryResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	if s.cacheRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("repository cache not configured"))
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
	client, err := s.resolveProviderClient(ctx, ws, binding)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	bindingKey := repoCacheBindingKey(binding)
	maxFileBytes, maxCacheBytes := s.cacheLimits()
	if maxFileBytes <= 0 {
		maxFileBytes = defaultMaxFileBytes
	}
	if maxCacheBytes <= 0 {
		maxCacheBytes = defaultMaxWorkspaceCache
	}

	headSHA, err := client.GetBranchHead(ctx, binding.GetBranch())
	if err != nil {
		// Provider failure: enter DEGRADED state but keep active agents.
		syncErr := "sync failed: " + providerErrDetail(err)
		binding.LastSyncError = syncErr
		if errors.Is(err, gitprovider.ErrUnauthorized) || errors.Is(err, gitprovider.ErrForbidden) {
			binding.Status = &agentsv1.RepoBindingStatus{
				State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_DEGRADED,
				Error: syncErr,
			}
		}
		if _, putErr := s.repo.Put(ctx, ws, binding); putErr != nil {
			logger.Error("persist sync error failed", "workspace_id", ws, "err", putErr)
		}
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New(syncErr))
	}

	// Non-fast-forward detection: if we have a known baseline and the new
	// HEAD is not a descendant, the branch was force-pushed.
	baseline := binding.GetActiveCommitSha()
	if baseline == "" {
		baseline = binding.GetObservedCommitSha()
	}
	if baseline != "" && baseline != headSHA {
		cmp, cmpErr := client.CompareCommits(ctx, baseline, headSHA)
		if cmpErr == nil && cmp.Status == "diverged" {
			logger.Warn("non-fast-forward detected", "workspace_id", ws,
				"baseline", baseline, "new_head", headSHA)
			binding.Status = &agentsv1.RepoBindingStatus{
				State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_DIVERGED,
				Error: fmt.Sprintf("non-fast-forward: new HEAD %s is not a descendant of %s", headSHA[:minLen(headSHA, 12)], baseline[:minLen(baseline, 12)]),
			}
			binding.ObservedCommitSha = headSHA
			binding.LastSyncError = binding.Status.GetError()
			if _, putErr := s.repo.Put(ctx, ws, binding); putErr != nil {
				logger.Error("persist diverged status failed", "workspace_id", ws, "err", putErr)
			}
			return connect.NewResponse(&agentsv1.SyncWorkspaceRepositoryResponse{
				Binding: binding,
			}), nil
		}
	}

	// Idempotent: skip if already synced to this commit.
	if binding.GetObservedCommitSha() == headSHA {
		metadata, cacheErr := s.cacheRepo.GetMetadata(ctx, ws)
		if cacheErr == nil && metadata.BindingKey == bindingKey && metadata.CommitSHA == headSHA {
			logger.Info("sync skipped (idempotent)", "workspace_id", ws, "sha", headSHA)
			return connect.NewResponse(&agentsv1.SyncWorkspaceRepositoryResponse{
				Binding:       binding,
				EntriesSynced: 0,
			}), nil
		}
	}

	treePath := binding.GetRootPath()
	// Use the pinned commit SHA for tree and blob reads so the cache
	// cannot contain mixed revisions if the branch advances mid-sync.
	treeEntries, err := client.GetTree(ctx, headSHA, treePath)
	if err != nil {
		syncErr := "sync failed: cannot read tree: " + providerErrDetail(err)
		binding.LastSyncError = syncErr
		binding.ObservedCommitSha = headSHA
		if _, putErr := s.repo.Put(ctx, ws, binding); putErr != nil {
			logger.Error("persist sync error failed", "workspace_id", ws, "err", putErr)
		}
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New(syncErr))
	}

	// Strip root_path prefix so cached entries are relative to the
	// binding's root, making browser listings work regardless of root_path.
	rootPrefix := strings.TrimRight(treePath, "/")
	if rootPrefix != "" {
		rootPrefix += "/"
	}

	var cacheEntries []*agentsv1.RepoCacheEntry
	var blobs []repocache.CachedBlob
	var totalBytes int64
	knownAgentIDs := make(map[string]struct{})
	if s.agentRepo != nil {
		agents, listErr := s.agentRepo.ListAgents(ctx, ws)
		if listErr != nil {
			return nil, connectx.InternalWith(listErr)
		}
		for _, agent := range agents {
			if agentID := agent.GetAgentId(); agentID != "" {
				knownAgentIDs[agentID] = struct{}{}
			}
		}
	}

	for _, te := range treeEntries {
		// Full repo-relative path (used for blob fetching).
		fullPath := te.Path
		// Cache-relative path (root_path stripped).
		cachePath := te.Path
		if rootPrefix != "" {
			cachePath = strings.TrimPrefix(te.Path, rootPrefix)
			if cachePath == te.Path {
				// Entry is the root_path directory itself; skip it.
				continue
			}
		}

		kind := treeEntryKindToProto(te.Kind)

		if err := validateCachePath(cachePath); err != nil {
			logger.Warn("skipping invalid path", "path", cachePath, "reason", err.Error())
			continue
		}

		entry := &agentsv1.RepoCacheEntry{
			Path:        cachePath,
			Kind:        kind,
			Size:        te.Size,
			ContentHash: te.SHA,
			Claimed:     isClaimedAgentPath(cachePath, knownAgentIDs),
		}
		cacheEntries = append(cacheEntries, entry)

		if kind != agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_FILE {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(cachePath), ".md") {
			continue
		}
		// Pre-check with reported size when available (GitHub provides it;
		// GitLab may not). Actual size is re-checked after download.
		if te.Size > 0 && te.Size > maxFileBytes {
			logger.Warn("skipping oversized file", "path", cachePath, "size", te.Size, "max", maxFileBytes)
			continue
		}

		data, err := client.GetBlob(ctx, headSHA, fullPath)
		if err != nil {
			logger.Warn("skipping unreadable blob", "path", cachePath, "err", err)
			continue
		}
		if !utf8.Valid(data) {
			logger.Warn("skipping non-UTF-8 file", "path", cachePath)
			continue
		}

		actualSize := int64(len(data))
		if actualSize > maxFileBytes {
			logger.Warn("skipping oversized file (post-download)", "path", cachePath, "size", actualSize, "max", maxFileBytes)
			continue
		}
		if totalBytes+actualSize > maxCacheBytes {
			syncErr := fmt.Sprintf("sync failed: workspace cache content exceeds configured limit of %d bytes", maxCacheBytes)
			binding.ObservedCommitSha = headSHA
			binding.LastSyncError = syncErr
			if _, putErr := s.repo.Put(ctx, ws, binding); putErr != nil {
				logger.Error("persist sync limit error failed", "workspace_id", ws, "err", putErr)
			}
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New(syncErr))
		}

		h := sha256.Sum256(data)
		entry.ContentHash = hex.EncodeToString(h[:])
		entry.Size = actualSize
		totalBytes += actualSize

		blobs = append(blobs, repocache.CachedBlob{
			Path:    cachePath,
			Content: data,
		})
	}

	if err := s.cacheRepo.PutSnapshot(ctx, ws, repocache.SnapshotMetadata{
		BindingKey: bindingKey,
		CommitSHA:  headSHA,
	}, cacheEntries, blobs); err != nil {
		logger.Error("persist cache snapshot failed", "workspace_id", ws, "err", err)
		return nil, connectx.InternalWith(err)
	}

	now := time.Now().UTC()
	binding.ObservedCommitSha = headSHA
	binding.LastSyncedAt = timestamppb.New(now)
	binding.LastSyncError = ""
	stored, err := s.repo.Put(ctx, ws, binding)
	if err != nil {
		logger.Error("persist binding after sync failed", "workspace_id", ws, "err", err)
		return nil, mapRepoBindingErr(err)
	}

	logger.Info("repo sync completed", "workspace_id", ws, "sha", headSHA, "entries", len(cacheEntries), "blobs", len(blobs))

	// Attempt publication of the newly synced cache.
	published, pubErrors, pubErr := s.publishActiveRevision(ctx, ws)
	if pubErr != nil {
		logger.Error("publication after sync failed", "workspace_id", ws, "err", pubErr)
	}

	// Re-read binding for final state.
	final, err := s.repo.Get(ctx, ws)
	if err != nil {
		final = stored
	}

	return connect.NewResponse(&agentsv1.SyncWorkspaceRepositoryResponse{
		Binding:           final,
		EntriesSynced:     int32(len(cacheEntries)),
		Published:         published,
		PublicationErrors: pubErrors,
	}), nil
}

func (s *RepoBindingServiceServer) GetRepositorySyncStatus(ctx context.Context, _ *connect.Request[agentsv1.GetRepositorySyncStatusRequest]) (*connect.Response[agentsv1.GetRepositorySyncStatusResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}
	return connect.NewResponse(&agentsv1.GetRepositorySyncStatusResponse{Binding: binding}), nil
}

func (s *RepoBindingServiceServer) ListRepositoryEntries(ctx context.Context, req *connect.Request[agentsv1.ListRepositoryEntriesRequest]) (*connect.Response[agentsv1.ListRepositoryEntriesResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	if s.cacheRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("repository cache not configured"))
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	// Require an active binding so stale caches from deleted bindings are
	// inaccessible.
	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}
	dirPath := strings.TrimRight(req.Msg.GetPath(), "/")
	if err := validateCachePath(dirPath); dirPath != "" && err != nil {
		return nil, connectx.InvalidArgument("path", err.Error())
	}
	metadata, err := s.cacheRepo.GetMetadata(ctx, ws)
	if err != nil {
		if errors.Is(err, repocache.ErrNotFound) {
			return nil, connectx.NotFound("no cached repository data; trigger a sync first")
		}
		return nil, connectx.InternalWith(err)
	}
	if metadata.BindingKey != repoCacheBindingKey(binding) {
		return nil, connectx.NotFound("no cached repository data for the current binding; trigger a sync first")
	}
	entries, err := s.cacheRepo.ListEntries(ctx, ws, metadata.SnapshotID, dirPath)
	if err != nil {
		if errors.Is(err, repocache.ErrNotFound) {
			return nil, connectx.NotFound("no cached repository data; trigger a sync first")
		}
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListRepositoryEntriesResponse{
		CommitSha:         metadata.CommitSHA,
		Entries:           entries,
		ObservedCommitSha: binding.GetObservedCommitSha(),
		ActiveCommitSha:   binding.GetActiveCommitSha(),
	}), nil
}

func (s *RepoBindingServiceServer) GetRepositoryFile(ctx context.Context, req *connect.Request[agentsv1.GetRepositoryFileRequest]) (*connect.Response[agentsv1.GetRepositoryFileResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	if s.cacheRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("repository cache not configured"))
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}
	filePath := req.Msg.GetPath()
	if filePath == "" {
		return nil, connectx.RequiredArgument("path")
	}
	if err := validateCachePath(filePath); err != nil {
		return nil, connectx.InvalidArgument("path", err.Error())
	}
	metadata, err := s.cacheRepo.GetMetadata(ctx, ws)
	if err != nil {
		if errors.Is(err, repocache.ErrNotFound) {
			return nil, connectx.NotFound("no cached repository data; trigger a sync first")
		}
		return nil, connectx.InternalWith(err)
	}
	if metadata.BindingKey != repoCacheBindingKey(binding) {
		return nil, connectx.NotFound("no cached repository data for the current binding; trigger a sync first")
	}
	entry, err := s.cacheRepo.GetEntry(ctx, ws, metadata.SnapshotID, filePath)
	if err != nil {
		if errors.Is(err, repocache.ErrNotFound) {
			return nil, connectx.NotFound("file not found in cache")
		}
		return nil, connectx.InternalWith(err)
	}
	content, err := s.cacheRepo.GetBlob(ctx, ws, metadata.SnapshotID, filePath)
	if err != nil {
		if errors.Is(err, repocache.ErrNotFound) {
			return nil, connectx.NotFound("file content not cached (may exceed size limit)")
		}
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.GetRepositoryFileResponse{
		CommitSha:         metadata.CommitSHA,
		Entry:             entry,
		Content:           string(content),
		ObservedCommitSha: binding.GetObservedCommitSha(),
		ActiveCommitSha:   binding.GetActiveCommitSha(),
	}), nil
}

// ── Helpers ─────────────────────────────────────────────────────────────

func treeEntryKindToProto(k gitprovider.TreeEntryKind) agentsv1.RepoCacheEntryKind {
	switch k {
	case gitprovider.TreeEntryFile:
		return agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_FILE
	case gitprovider.TreeEntryDirectory:
		return agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_DIRECTORY
	case gitprovider.TreeEntrySymlink:
		return agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_SYMLINK
	case gitprovider.TreeEntrySubmodule:
		return agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_SUBMODULE
	default:
		return agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_UNSPECIFIED
	}
}

func minLen(s string, n int) int {
	if len(s) < n {
		return len(s)
	}
	return n
}

// validateCachePath rejects dangerous path patterns: absolute paths,
// traversal segments, backslashes, and NUL bytes.
func validateCachePath(p string) error {
	if p == "" {
		return nil
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return errors.New("must be a relative path")
	}
	if strings.ContainsRune(p, 0) {
		return errors.New("path contains NUL byte")
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return errors.New("must not contain traversal segments")
		}
	}
	return nil
}

// ── Publication pipeline (issue #216) ───────────────────────────────────

// publishActiveRevision validates the current observed cache, writes an
// Agent Content snapshot, sets active_commit_sha, and reloads the runner.
// On validation failure, the active revision is unchanged.
func (s *RepoBindingServiceServer) publishActiveRevision(ctx context.Context, ws string) (published bool, validationErrors []string, err error) {
	if s.contentRepo == nil || s.cacheRepo == nil {
		return false, nil, nil
	}

	mu := s.publishMutex(ws)
	mu.Lock()
	defer mu.Unlock()

	logger := log.FromContext(ctx)

	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return false, nil, err
	}

	observedSHA := binding.GetObservedCommitSha()
	if observedSHA == "" {
		return false, nil, nil
	}

	// Idempotent: already published this revision.
	if binding.GetActiveCommitSha() == observedSHA {
		return true, nil, nil
	}

	metadata, err := s.cacheRepo.GetMetadata(ctx, ws)
	if err != nil {
		return false, nil, fmt.Errorf("cache metadata: %w", err)
	}
	if metadata.CommitSHA != observedSHA {
		return false, nil, fmt.Errorf("cache SHA %q does not match observed %q", metadata.CommitSHA, observedSHA)
	}

	agents, err := s.agentRepo.ListAgents(ctx, ws)
	if err != nil {
		return false, nil, fmt.Errorf("list agents: %w", err)
	}
	var agentIDs []string
	for _, a := range agents {
		if id := a.GetAgentId(); id != "" {
			agentIDs = append(agentIDs, id)
		}
	}

	blobReader := func(p string) ([]byte, bool) {
		data, err := s.cacheRepo.GetBlob(ctx, ws, metadata.SnapshotID, p)
		if err != nil {
			return nil, false
		}
		return data, true
	}

	parsed := agentcontent.Parse(blobReader, agentIDs)
	if len(parsed.Content) == 0 {
		logger.Info("publication skipped: no agent content in cache", "workspace_id", ws)
		binding.LastPublicationError = ""
		binding.PublicationErrors = nil
		_, _ = s.repo.Put(ctx, ws, binding)
		return true, nil, nil
	}

	validationErrs := agentcontent.Validate(parsed.Content, agents)
	if len(validationErrs) > 0 {
		errStrs := make([]string, len(validationErrs))
		for i, ve := range validationErrs {
			errStrs[i] = ve.Error()
		}
		logger.Warn("publication failed: validation errors", "workspace_id", ws, "errors", errStrs)
		binding.LastPublicationError = fmt.Sprintf("%d validation error(s)", len(validationErrs))
		binding.PublicationErrors = errStrs
		if _, putErr := s.repo.Put(ctx, ws, binding); putErr != nil {
			logger.Error("persist publication errors failed", "workspace_id", ws, "err", putErr)
		}
		return false, errStrs, nil
	}

	snapshot := agentcontent.Snapshot{
		CommitSHA: observedSHA,
		Entries:   parsed.Content,
	}
	if err := s.contentRepo.PutSnapshot(ctx, ws, snapshot); err != nil {
		return false, nil, fmt.Errorf("store content snapshot: %w", err)
	}

	binding.ActiveCommitSha = observedSHA
	binding.LastPublicationError = ""
	binding.PublicationErrors = nil
	if _, err := s.repo.Put(ctx, ws, binding); err != nil {
		return false, nil, fmt.Errorf("advance active revision: %w", err)
	}

	if s.configRuntime != nil {
		if err := s.configRuntime.ReloadRunner(ctx); err != nil {
			logger.Error("runner reload after publication failed", "workspace_id", ws, "err", err)
			return true, nil, nil
		}
	}

	logger.Info("publication succeeded", "workspace_id", ws, "active_commit_sha", observedSHA,
		"agents_published", len(parsed.Content))
	return true, nil, nil
}

func (s *RepoBindingServiceServer) PublishWorkspaceRepository(ctx context.Context, _ *connect.Request[agentsv1.PublishWorkspaceRepositoryRequest]) (*connect.Response[agentsv1.PublishWorkspaceRepositoryResponse], error) {
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

	_, validationErrors, pubErr := s.publishActiveRevision(ctx, ws)
	if pubErr != nil {
		return nil, connectx.InternalWith(pubErr)
	}

	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}
	return connect.NewResponse(&agentsv1.PublishWorkspaceRepositoryResponse{
		Binding:          binding,
		ValidationErrors: validationErrors,
	}), nil
}

// ── Webhook secret management (issue #216) ──────────────────────────────

func (s *RepoBindingServiceServer) ConfigureWebhookSecret(ctx context.Context, _ *connect.Request[agentsv1.ConfigureWebhookSecretRequest]) (*connect.Response[agentsv1.ConfigureWebhookSecretResponse], error) {
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

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, connectx.Internal("failed to generate webhook secret")
	}
	secretHex := hex.EncodeToString(secret)

	cipher, err := s.cipher()
	if err != nil {
		return nil, err
	}
	ciphertext, encErr := cipher.Encrypt([]byte(secretHex))
	if encErr != nil {
		return nil, connectx.Internal("failed to encrypt webhook secret")
	}

	if err := s.repo.SetWebhookSecret(ctx, ws, ciphertext); err != nil {
		return nil, mapRepoBindingErr(err)
	}

	binding.WebhookSecretSet = true
	stored, err := s.repo.Put(ctx, ws, binding)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}

	callbackURL := "/api/webhooks/repository/" + ws
	if s.webhookBaseURL != nil {
		if base := s.webhookBaseURL(); base != "" {
			callbackURL = strings.TrimRight(base, "/") + callbackURL
		}
	}

	log.FromContext(ctx).Info("webhook secret configured", "workspace_id", ws)
	return connect.NewResponse(&agentsv1.ConfigureWebhookSecretResponse{
		Binding:       stored,
		WebhookSecret: secretHex,
		CallbackUrl:   callbackURL,
	}), nil
}

// VerifyWebhookSignature checks a GitHub-style HMAC-SHA256 signature or a
// GitLab-style secret token header. Returns true if the signature is valid.
func (s *RepoBindingServiceServer) VerifyWebhookSignature(ctx context.Context, ws string, body []byte, signatureHeader, tokenHeader string) bool {
	if s.repo == nil {
		return false
	}
	ciphertext, err := s.repo.GetWebhookSecret(ctx, ws)
	if err != nil {
		return false
	}
	cipher, err := s.cipher()
	if err != nil {
		return false
	}
	secret, err := cipher.Decrypt(ciphertext)
	if err != nil {
		return false
	}
	secretStr := string(secret)

	// GitHub: X-Hub-Signature-256 = sha256=<hex>
	if signatureHeader != "" {
		expected := "sha256=" + computeHMACSHA256(secretStr, body)
		return hmac.Equal([]byte(expected), []byte(signatureHeader))
	}

	// GitLab: X-Gitlab-Token = <secret>
	if tokenHeader != "" {
		return hmac.Equal([]byte(secretStr), []byte(tokenHeader))
	}
	return false
}

func computeHMACSHA256(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// TriggerSyncAndPublish runs a full sync + publication for the workspace.
// Used by webhook handler and periodic reconciliation. The context is
// elevated to admin so the role check is bypassed.
func (s *RepoBindingServiceServer) TriggerSyncAndPublish(ctx context.Context, ws string) error {
	ctx = auth.WithAdmin(withWorkspace(ctx, ws))
	_, err := s.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	return err
}

// ── Agent Content editing (issue #217) ──────────────────────────────────

const (
	defaultMaxContentFileBytes = 256 * 1024 // 256 KiB per file
)

// CommitAgentContent applies a changeset of PUT and DELETE file operations
// to managed Agent Content paths and produces a single Git commit. In
// DIRECT_COMMIT mode the commit lands on the bound branch; in
// CHANGE_REQUEST mode a PR/MR is opened instead.
func (s *RepoBindingServiceServer) CommitAgentContent(ctx context.Context, req *connect.Request[agentsv1.CommitAgentContentRequest]) (*connect.Response[agentsv1.CommitAgentContentResponse], error) {
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

	actions := req.Msg.GetActions()
	if len(actions) == 0 {
		return nil, connectx.RequiredArgument("actions")
	}

	// Validate and normalize all actions.
	maxFileBytes, _ := s.cacheLimits()
	if maxFileBytes <= 0 {
		maxFileBytes = defaultMaxContentFileBytes
	}
	if err := validateContentActions(actions, maxFileBytes); err != nil {
		return nil, err
	}

	// Simulate the resulting content to validate before committing.
	agents, err := s.agentRepo.ListAgents(ctx, ws)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	validationErrors := s.simulateAndValidate(ctx, ws, actions, agents)
	if len(validationErrors) > 0 {
		return connect.NewResponse(&agentsv1.CommitAgentContentResponse{
			Binding:          binding,
			ValidationErrors: validationErrors,
		}), nil
	}

	client, err := s.resolveProviderClient(ctx, ws, binding)
	if err != nil {
		return nil, err
	}

	// Last-write-wins: always commit against the latest HEAD.
	headSHA, err := client.GetBranchHead(ctx, binding.GetBranch())
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("cannot read branch HEAD: %s", providerErrDetail(err)))
	}

	// Build provider file actions with root_path-prefixed paths.
	rootPrefix := binding.GetRootPath()
	if rootPrefix != "" {
		rootPrefix = strings.TrimRight(rootPrefix, "/") + "/"
	}
	providerActions := make([]gitprovider.FileAction, len(actions))
	for i, a := range actions {
		fullPath := rootPrefix + a.GetPath()
		providerActions[i] = gitprovider.FileAction{
			Path:    fullPath,
			Delete:  a.GetOperation() == agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_DELETE,
			Content: []byte(a.GetContent()),
		}
	}

	// Build commit message with audit metadata.
	commitMsg := req.Msg.GetMessage()
	if commitMsg == "" {
		commitMsg = "Update Agent Content from Butter"
	}
	actor := "unknown"
	if user, ok := auth.UserFromContext(ctx); ok {
		actor = user.GetId()
	}
	auditTrailer := fmt.Sprintf("\nButter-Actor: %s\nButter-Workspace: %s\nButter-Operation: commit\nButter-Base-SHA: %s\nButter-Parent-SHA: %s",
		actor, ws, req.Msg.GetBaseRevision(), headSHA)
	commitMsg += auditTrailer

	logger := log.FromContext(ctx)

	isChangeRequest := binding.GetWriteMode() == agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_CHANGE_REQUEST
	var commitResult *gitprovider.CommitResult
	var crURL string

	if isChangeRequest {
		branchName := fmt.Sprintf("butter/content-%s-%d", ws, time.Now().Unix())
		if err := client.CreateBranch(ctx, branchName, headSHA); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create branch: %s", err))
		}
		commitResult, err = client.CreateCommit(ctx, branchName, headSHA, commitMsg, providerActions)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create commit: %s", err))
		}
		cr, crErr := client.CreateChangeRequest(ctx, branchName, binding.GetBranch(),
			"Agent Content Update", commitMsg)
		if crErr != nil {
			logger.Error("created commit but failed to open change request",
				"workspace_id", ws, "branch", branchName, "err", crErr)
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create change request: %s", crErr))
		}
		crURL = cr.URL
	} else {
		commitResult, err = client.CreateCommit(ctx, binding.GetBranch(), headSHA, commitMsg, providerActions)
		if err != nil {
			if errors.Is(err, gitprovider.ErrConflict) {
				return nil, connect.NewError(connect.CodeAborted, errors.New("branch HEAD moved; retry the operation"))
			}
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create commit: %s", err))
		}
	}

	logger.Info("agent content committed", "workspace_id", ws, "commit_sha", commitResult.SHA,
		"actions", len(actions), "change_request", isChangeRequest)

	// Trigger sync + publish so the runner picks up the new content.
	if !isChangeRequest {
		if syncErr := s.TriggerSyncAndPublish(ctx, ws); syncErr != nil {
			logger.Error("sync after content commit failed", "workspace_id", ws, "err", syncErr)
		}
	}

	final, err := s.repo.Get(ctx, ws)
	if err != nil {
		final = binding
	}
	return connect.NewResponse(&agentsv1.CommitAgentContentResponse{
		Binding:          final,
		CommitSha:        commitResult.SHA,
		ChangeRequestUrl: crURL,
	}), nil
}

// RollbackAgentContent restores managed Agent Content from a previously
// published revision in a new Git commit.
func (s *RepoBindingServiceServer) RollbackAgentContent(ctx context.Context, req *connect.Request[agentsv1.RollbackAgentContentRequest]) (*connect.Response[agentsv1.RollbackAgentContentResponse], error) {
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

	targetSHA := strings.TrimSpace(req.Msg.GetTargetCommitSha())
	if targetSHA == "" {
		return nil, connectx.RequiredArgument("target_commit_sha")
	}

	client, err := s.resolveProviderClient(ctx, ws, binding)
	if err != nil {
		return nil, err
	}

	// Read managed content from the target revision.
	treePath := binding.GetRootPath()
	targetEntries, err := client.GetTree(ctx, targetSHA, treePath)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("cannot read tree at %s: %s", targetSHA[:minLen(targetSHA, 12)], providerErrDetail(err)))
	}

	rootPrefix := strings.TrimRight(treePath, "/")
	if rootPrefix != "" {
		rootPrefix += "/"
	}

	// Get latest HEAD for last-write-wins.
	headSHA, err := client.GetBranchHead(ctx, binding.GetBranch())
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("cannot read branch HEAD: %s", providerErrDetail(err)))
	}

	// Read current managed tree at HEAD for DELETE detection.
	currentEntries, currentTreeErr := client.GetTree(ctx, headSHA, treePath)
	if currentTreeErr != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("cannot read current tree at %s: %s", headSHA[:minLen(headSHA, 12)], providerErrDetail(currentTreeErr)))
	}

	// Build the set of managed file paths from the target revision (PUT)
	// and from the current HEAD (DELETE files absent from target).
	maxFileBytes, _ := s.cacheLimits()
	if maxFileBytes <= 0 {
		maxFileBytes = defaultMaxContentFileBytes
	}

	var providerActions []gitprovider.FileAction
	targetPaths := make(map[string]bool)
	for _, te := range targetEntries {
		if te.Kind != gitprovider.TreeEntryFile {
			continue
		}
		cachePath := te.Path
		if rootPrefix != "" {
			cachePath = strings.TrimPrefix(te.Path, rootPrefix)
		}
		if !strings.HasPrefix(cachePath, "agents/") || !strings.HasSuffix(strings.ToLower(cachePath), ".md") {
			continue
		}
		data, blobErr := client.GetBlob(ctx, targetSHA, te.Path)
		if blobErr != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("cannot read %s at %s: %s", cachePath, targetSHA[:minLen(targetSHA, 12)], providerErrDetail(blobErr)))
		}
		if int64(len(data)) > maxFileBytes {
			return nil, connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("file %s exceeds maximum size of %d bytes", cachePath, maxFileBytes))
		}
		targetPaths[te.Path] = true
		providerActions = append(providerActions, gitprovider.FileAction{
			Path:    te.Path,
			Content: data,
		})
	}

	// The target revision must contain at least one managed Agent Content
	// file; a rollback to an empty revision is rejected.
	if len(targetPaths) == 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("target revision has no managed Agent Content"))
	}

	// Validate the target revision's content before committing (criterion 4).
	agents, agentErr := s.agentRepo.ListAgents(ctx, ws)
	if agentErr != nil {
		return nil, connectx.InternalWith(agentErr)
	}
	rollbackContent := make(map[string]agentcontent.AgentContent)
	for _, a := range providerActions {
		if a.Delete {
			continue
		}
		p := a.Path
		if rootPrefix != "" {
			p = strings.TrimPrefix(a.Path, rootPrefix)
		}
		parts := strings.Split(p, "/")
		if len(parts) < 3 || parts[0] != "agents" {
			continue
		}
		agentID := parts[1]
		fileName := parts[len(parts)-1]
		existing := rollbackContent[agentID]
		existing.AgentID = agentID
		switch fileName {
		case "description.md":
			existing.Description = strings.TrimSpace(string(a.Content))
		case "prompt.md":
			existing.Instruction = strings.TrimSpace(string(a.Content))
		case "global-prompt.md":
			existing.GlobalInstruction = strings.TrimSpace(string(a.Content))
		}
		rollbackContent[agentID] = existing
	}
	validationErrs := agentcontent.Validate(rollbackContent, agents)
	if len(validationErrs) > 0 {
		errStrs := make([]string, len(validationErrs))
		for i, ve := range validationErrs {
			errStrs[i] = ve.Error()
		}
		return connect.NewResponse(&agentsv1.RollbackAgentContentResponse{
			Binding:          binding,
			ValidationErrors: errStrs,
		}), nil
	}

	// DELETE files present at HEAD but absent from the target revision.
	for _, ce := range currentEntries {
		if ce.Kind != gitprovider.TreeEntryFile {
			continue
		}
		cachePath := ce.Path
		if rootPrefix != "" {
			cachePath = strings.TrimPrefix(ce.Path, rootPrefix)
		}
		if !strings.HasPrefix(cachePath, "agents/") || !strings.HasSuffix(strings.ToLower(cachePath), ".md") {
			continue
		}
		if !targetPaths[ce.Path] {
			providerActions = append(providerActions, gitprovider.FileAction{
				Path:   ce.Path,
				Delete: true,
			})
		}
	}

	// Build rollback commit message with audit metadata.
	actor := "unknown"
	if user, ok := auth.UserFromContext(ctx); ok {
		actor = user.GetId()
	}
	commitMsg := fmt.Sprintf("Rollback Agent Content to %s\n\nButter-Actor: %s\nButter-Workspace: %s\nButter-Operation: rollback\nButter-Target-SHA: %s\nButter-Parent-SHA: %s",
		targetSHA[:minLen(targetSHA, 12)], actor, ws, targetSHA, headSHA)

	logger := log.FromContext(ctx)
	isChangeRequest := binding.GetWriteMode() == agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_CHANGE_REQUEST
	var commitResult *gitprovider.CommitResult
	var crURL string

	if isChangeRequest {
		branchName := fmt.Sprintf("butter/rollback-%s-%d", ws, time.Now().Unix())
		if err := client.CreateBranch(ctx, branchName, headSHA); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create branch: %s", err))
		}
		commitResult, err = client.CreateCommit(ctx, branchName, headSHA, commitMsg, providerActions)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create commit: %s", err))
		}
		cr, crErr := client.CreateChangeRequest(ctx, branchName, binding.GetBranch(),
			"Rollback Agent Content to "+targetSHA[:minLen(targetSHA, 12)], commitMsg)
		if crErr != nil {
			logger.Error("rollback commit created but failed to open change request",
				"workspace_id", ws, "err", crErr)
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create change request: %s", crErr))
		}
		crURL = cr.URL
	} else {
		commitResult, err = client.CreateCommit(ctx, binding.GetBranch(), headSHA, commitMsg, providerActions)
		if err != nil {
			if errors.Is(err, gitprovider.ErrConflict) {
				return nil, connect.NewError(connect.CodeAborted, errors.New("branch HEAD moved; retry the operation"))
			}
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create commit: %s", err))
		}
	}

	logger.Info("agent content rolled back", "workspace_id", ws, "commit_sha", commitResult.SHA,
		"target_sha", targetSHA, "change_request", isChangeRequest)

	// Trigger sync + publish.
	if !isChangeRequest {
		if syncErr := s.TriggerSyncAndPublish(ctx, ws); syncErr != nil {
			logger.Error("sync after rollback failed", "workspace_id", ws, "err", syncErr)
		}
	}

	final, err := s.repo.Get(ctx, ws)
	if err != nil {
		final = binding
	}
	return connect.NewResponse(&agentsv1.RollbackAgentContentResponse{
		Binding:          final,
		CommitSha:        commitResult.SHA,
		ChangeRequestUrl: crURL,
	}), nil
}

// validateContentActions validates all changeset actions: paths must be
// within the managed agents subtree, content must be valid UTF-8 Markdown,
// and file sizes must be within limits.
func validateContentActions(actions []*agentsv1.ContentFileAction, maxFileBytes int64) error {
	for _, a := range actions {
		p := strings.TrimSpace(a.GetPath())
		if p == "" {
			return connectx.RequiredArgument("actions[].path")
		}
		if err := validateCachePath(p); err != nil {
			return connectx.InvalidArgument("actions[].path", err.Error())
		}
		if !strings.HasPrefix(p, "agents/") {
			return connectx.InvalidArgument("actions[].path", "must be within the agents/ subtree")
		}
		if !strings.HasSuffix(strings.ToLower(p), ".md") {
			return connectx.InvalidArgument("actions[].path", "must be a Markdown (.md) file")
		}
		op := a.GetOperation()
		if op == agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_UNSPECIFIED {
			return connectx.RequiredArgument("actions[].operation")
		}
		if op == agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT {
			content := a.GetContent()
			if !utf8.ValidString(content) {
				return connectx.InvalidArgument("actions[].content", "must be valid UTF-8")
			}
			if int64(len(content)) > maxFileBytes {
				return connectx.InvalidArgument("actions[].content",
					fmt.Sprintf("exceeds maximum file size of %d bytes", maxFileBytes))
			}
		}
	}
	return nil
}

// simulateAndValidate predicts the resulting agent content after applying
// the changeset and validates it. Returns validation error strings (empty
// means OK to commit).
func (s *RepoBindingServiceServer) simulateAndValidate(ctx context.Context, ws string, actions []*agentsv1.ContentFileAction, agents []*agentsv1.Agent) []string {
	// Start from the current active content snapshot if available.
	simulated := make(map[string]agentcontent.AgentContent)
	if s.contentRepo != nil {
		if snap, err := s.contentRepo.GetSnapshot(ctx, ws); err == nil {
			for k, v := range snap.Entries {
				simulated[k] = v
			}
		}
	}

	// Apply actions.
	for _, a := range actions {
		p := a.GetPath()
		parts := strings.Split(p, "/")
		if len(parts) < 3 || parts[0] != "agents" {
			continue
		}
		agentID := parts[1]
		fileName := parts[len(parts)-1]

		if a.GetOperation() == agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_DELETE {
			if existing, ok := simulated[agentID]; ok {
				switch fileName {
				case "description.md":
					existing.Description = ""
				case "prompt.md":
					existing.Instruction = ""
				case "global-prompt.md":
					existing.GlobalInstruction = ""
				}
				if existing.Description == "" && existing.Instruction == "" && existing.GlobalInstruction == "" {
					delete(simulated, agentID)
				} else {
					simulated[agentID] = existing
				}
			}
			continue
		}

		content := strings.TrimSpace(a.GetContent())
		existing := simulated[agentID]
		existing.AgentID = agentID
		switch fileName {
		case "description.md":
			existing.Description = content
		case "prompt.md":
			existing.Instruction = content
		case "global-prompt.md":
			existing.GlobalInstruction = content
		}
		simulated[agentID] = existing
	}

	validationErrs := agentcontent.Validate(simulated, agents)
	if len(validationErrs) == 0 {
		return nil
	}
	errStrs := make([]string, len(validationErrs))
	for i, ve := range validationErrs {
		errStrs[i] = ve.Error()
	}
	return errStrs
}

// ── Baseline acceptance (issue #216) ────────────────────────────────────

func (s *RepoBindingServiceServer) AcceptRepositoryBaseline(ctx context.Context, _ *connect.Request[agentsv1.AcceptRepositoryBaselineRequest]) (*connect.Response[agentsv1.AcceptRepositoryBaselineResponse], error) {
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
	if binding.GetStatus().GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_DIVERGED {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("binding is not in DIVERGED state"))
	}

	// Reset state to OK so the sync proceeds normally.
	binding.Status = &agentsv1.RepoBindingStatus{
		State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_OK,
	}
	if _, err := s.repo.Put(ctx, ws, binding); err != nil {
		return nil, mapRepoBindingErr(err)
	}

	// Re-sync and publish.
	syncResp, syncErr := s.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if syncErr != nil {
		return nil, syncErr
	}
	return connect.NewResponse(&agentsv1.AcceptRepositoryBaselineResponse{
		Binding: syncResp.Msg.GetBinding(),
	}), nil
}
