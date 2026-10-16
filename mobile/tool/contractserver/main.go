// Command contractserver is a development-only HTTP fixture for the mobile
// binary Connect contract. It uses production services, codecs and auth
// middleware, with ephemeral in-memory persistence and no agent runtime.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/auth/provider"
	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo"
	agentopmemory "go.orx.me/apps/butter/internal/repo/agentop/memory"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	"go.orx.me/apps/butter/internal/repo/auth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	oauthhandoffmemory "go.orx.me/apps/butter/internal/repo/oauthhandoff/memory"
	oauthstatememory "go.orx.me/apps/butter/internal/repo/oauthstate/memory"
	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	"go.orx.me/apps/butter/internal/repo/workspace"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/service"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// Only the self-service auth boundary is implemented. Global administration
// RPCs are not registered, so the embedded interface is never used.
// Session copies keep the middleware's asynchronous TouchSession race-free.
type authStore struct {
	auth.Repository
	mu       sync.Mutex
	user     *agentsv1.User
	hash     string
	sessions map[string]auth.Session
}

func (s *authStore) FindUserByUsername(_ context.Context, username string) (*agentsv1.User, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if username != s.user.GetUsername() {
		return nil, "", auth.ErrUserNotFound
	}
	return proto.Clone(s.user).(*agentsv1.User), s.hash, nil
}

func (s *authStore) FindUserByID(_ context.Context, id string) (*agentsv1.User, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != s.user.GetId() {
		return nil, "", auth.ErrUserNotFound
	}
	return proto.Clone(s.user).(*agentsv1.User), s.hash, nil
}

// GetUser serves the OAuth handoff redemption. The profile and password writes
// below mutate the same user, so the read shares their lock.
func (s *authStore) GetUser(_ context.Context, id string) (*agentsv1.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != s.user.GetId() {
		return nil, auth.ErrUserNotFound
	}
	return proto.Clone(s.user).(*agentsv1.User), nil
}

// The fixture's one user is the account the fake provider identity is linked
// to, so OAuth sign-in never has to create users.
func (s *authStore) FindUserByExternalID(_ context.Context, providerName, externalID string) (*agentsv1.User, error) {
	if providerName != fakeProviderName || externalID != fakeExternalID {
		return nil, auth.ErrUserNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return proto.Clone(s.user).(*agentsv1.User), nil
}

func (s *authStore) UpdateUserProfile(_ context.Context, id, name string, avatarURL *string, at time.Time) (*agentsv1.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != s.user.GetId() {
		return nil, auth.ErrUserNotFound
	}
	s.user.DisplayName = name
	if avatarURL != nil {
		s.user.AvatarUrl = *avatarURL
	}
	s.user.UpdatedAt = timestamppb.New(at)
	return proto.Clone(s.user).(*agentsv1.User), nil
}

func (s *authStore) UpdateUserPassword(_ context.Context, id, hash string, at time.Time) (*agentsv1.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != s.user.GetId() {
		return nil, auth.ErrUserNotFound
	}
	s.hash = hash
	s.user.UpdatedAt = timestamppb.New(at)
	return proto.Clone(s.user).(*agentsv1.User), nil
}

func (s *authStore) CreateSession(_ context.Context, session *auth.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = *session
	return nil
}

func (s *authStore) LookupSession(_ context.Context, tokenHash string, now time.Time) (*auth.Session, *agentsv1.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, session := range s.sessions {
		if session.TokenHash == tokenHash && !session.Revoked && now.Before(session.ExpiresAt) {
			if s.user.GetDisabled() {
				return nil, nil, auth.ErrUserDisabled
			}
			return &session, proto.Clone(s.user).(*agentsv1.User), nil
		}
	}
	return nil, nil, auth.ErrSessionNotFound
}

func (s *authStore) TouchSession(_ context.Context, id string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok || session.Revoked {
		return auth.ErrSessionNotFound
	}
	session.LastUsedAt = at
	s.sessions[id] = session
	return nil
}

func (s *authStore) RevokeSession(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return auth.ErrSessionNotFound
	}
	delete(s.sessions, id)
	return nil
}

const (
	// Seeded Telegram references in Workspace alpha only. The real
	// ModelProviderService guard blocks removing these model aliases there,
	// and ignores them in beta (references are Workspace-scoped).
	guardedModelAlias      = "guarded-default"
	guardedSelectableAlias = "guarded-choice"
	guardedDestinationID   = "contract-destination"
)

const (
	fakeProviderName = "github"
	fakeProviderCode = "contract-provider-code"
	fakeExternalID   = "contract-external-id"
	// nativeReturnURI is the mobile App's development return URI.
	nativeReturnURI = "ee.orz.butter:/oauth/callback"
)

// fakeProvider stands in for the third-party OAuth provider, the only thing the
// fixture replaces on the OAuth path: AuthService, its state and handoff
// handling, and the HTTP middleware are the production ones.
type fakeProvider struct{}

func (fakeProvider) Name() string        { return fakeProviderName }
func (fakeProvider) DisplayName() string { return "GitHub" }
func (fakeProvider) AuthorizeURL(state string) string {
	return "https://provider.invalid/authorize?state=" + state
}
func (fakeProvider) Exchange(_ context.Context, code string) (*provider.Claims, error) {
	if code != fakeProviderCode {
		return nil, errors.New("fake provider: unknown code")
	}
	return &provider.Claims{Provider: fakeProviderName, ExternalID: fakeExternalID, Login: "mobile-test"}, nil
}

func fixture(ctx context.Context) (http.Handler, error) {
	handler, _, err := newFixture(ctx)
	return handler, err
}

// newFixture also returns the Daemon fixture, whose repositories and registry
// the Go tests inspect directly.
func newFixture(ctx context.Context) (http.Handler, *daemonFixture, error) {
	// These public credentials must never be used outside this loopback fixture.
	hash, err := bcrypt.GenerateFromPassword([]byte("contract-test-password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	authRepo := &authStore{
		user: &agentsv1.User{Id: "mobile-test", Username: "mobile-test", DisplayName: "Mobile Test", Role: "user"},
		hash: string(hash), sessions: make(map[string]auth.Session),
	}
	wsRepo := workspacememory.New()
	agentRepo := configmemory.New()
	for _, id := range []string{"alpha", "beta", chatWorkspace, boxWorkspace, mcpWorkspace, daemonWorkspace, daemonPeerWorkspace, filesWorkspace, filesPeerWorkspace, repoBindingWorkspace, repoBindingMemberWorkspace, sessionsWorkspace, sessionsPeerWS, cronWorkspace, memoryWorkspace, memoryMemberWorkspace, telegramWorkspace} {
		if _, err := wsRepo.CreateWorkspace(ctx, &agentsv1.Workspace{Id: id, Name: id, Slug: id}); err != nil {
			return nil, nil, err
		}
		role := "member"
		if id == daemonWorkspace || id == repoBindingWorkspace || id == sessionsWorkspace || id == sessionsPeerWS || id == cronWorkspace || id == memoryWorkspace || id == telegramWorkspace {
			// Owner role for Daemon token revocation and repo binding writes and memory writes.
			role = "owner"
		}
		if id == repoBindingMemberWorkspace || id == memoryMemberWorkspace {
			role = "member"
		}
		if _, err := wsRepo.AddMember(ctx, &agentsv1.WorkspaceMember{WorkspaceId: id, UserId: authRepo.user.GetId(), Role: role}); err != nil {
			return nil, nil, err
		}
		if id == chatWorkspace || id == boxWorkspace || id == mcpWorkspace || id == daemonWorkspace || id == daemonPeerWorkspace || id == filesWorkspace || id == filesPeerWorkspace || id == repoBindingWorkspace || id == sessionsWorkspace || id == sessionsPeerWS || id == cronWorkspace || id == memoryWorkspace || id == memoryMemberWorkspace || id == telegramWorkspace {
			// Chat's agent is the runnable one the chat fixture registers;
			// the box, MCP, Daemon and repo-binding Workspaces start with none,
			// their resources are created by the contract tests / repo fixture.
			continue
		}
		if _, err := agentRepo.CreateAgent(ctx, id, &agentsv1.Agent{AgentId: "shared", Name: id + "-agent"}); err != nil {
			return nil, nil, err
		}
	}
	// A Workspace the account is not a member of: every request naming it is
	// refused by the real membership check.
	if _, err := wsRepo.CreateWorkspace(ctx, &agentsv1.Workspace{Id: mcpDeniedWorkspace, Name: mcpDeniedWorkspace, Slug: mcpDeniedWorkspace}); err != nil {
		return nil, nil, err
	}
	// A real Telegram Destination routes to two model aliases in alpha, so the
	// real reference guard has something to refuse.
	telegramRepo := telegrammemory.New()
	if _, err := telegramRepo.CreateChannel(ctx, "alpha",
		&agentsv1.TelegramChannel{Id: "contract-channel", Key: "contract", Name: "contract", BotId: "424242"},
		telegramrepo.ChannelCredentials{BotToken: telegramrepo.Credential{Ciphertext: "unused", KeyID: "contract"}}); err != nil {
		return nil, nil, err
	}
	if _, err := telegramRepo.CreateDestination(ctx, "alpha", &agentsv1.TelegramDestination{
		Id: guardedDestinationID, Key: "contract-topic", ChannelId: "contract-channel", ChatId: "-1001",
		Config: &agentsv1.TelegramDestinationConfig{
			AgentId:          "shared",
			Model:            guardedModelAlias,
			SelectableModels: []string{guardedSelectableAlias},
		},
	}); err != nil {
		return nil, nil, err
	}
	modelProviderSvc := application.NewModelProviderServiceServer(agentRepo)
	modelProviderSvc.SetTelegramGuard(application.NewTelegramReferenceGuard(telegramRepo))
	authSvc := application.NewAuthServiceServer(authRepo, time.Hour)
	authSvc.SetWorkspaceRepo(wsRepo)
	providers := provider.NewRegistry()
	providers.Register(fakeProvider{})
	authSvc.SetProviderRegistry(providers)
	authSvc.SetOAuthStateRepo(oauthstatememory.New())
	authSvc.SetOAuthHandoffRepo(oauthhandoffmemory.New())
	authSvc.SetNativeReturnURIs([]string{nativeReturnURI})
	opts := connectx.HandlerOptions()
	_, authHandler := agentsv1connect.NewAuthServiceHandler(authSvc, opts...)
	_, wsHandler := agentsv1connect.NewWorkspaceServiceHandler(application.NewWorkspaceServiceServer(wsRepo), opts...)
	repoFx, err := newRepoBindingFixture(ctx, wsRepo, agentRepo)
	if err != nil {
		return nil, nil, err
	}
	// beta is bound but has no Active Revision: Agent Content remains
	// database-owned for existing Agent edits, including member writes.
	if _, err := repoFx.binding.Put(ctx, "beta", &agentsv1.WorkspaceRepoBinding{
		GitHostId: repoBindingGitHostID, Repository: "owner/repo", Branch: "main",
	}); err != nil {
		return nil, nil, err
	}
	agentSvc := application.NewAgentServiceServer(agentRepo)
	agentSvc.SetWorkspaceRepo(wsRepo)
	agentSvc.SetOperationRepo(agentopmemory.New())
	agentSvc.SetContentCoordinator(repoFx.svc)
	bx := newBoxFixture(agentRepo)
	agentSvc.SetButterBoxRepo(bx.repo)
	_, agentHandler := agentsv1connect.NewAgentServiceHandler(agentSvc, opts...)
	skillRepo := skillmemory.New()
	if _, err := skillRepo.Create(ctx, "alpha", &agentsv1.Skill{Name: "existing-skill"}, "fixture content"); err != nil {
		return nil, nil, err
	}
	if _, err := skillRepo.Create(ctx, filesWorkspace, &agentsv1.Skill{Name: filesSkill, Description: "Summarise a document"},
		"---\nname: "+filesSkill+"\ndescription: Summarise a document\n---\nKeep it short.\n"); err != nil {
		return nil, nil, err
	}
	_, skillHandler := agentsv1connect.NewSkillServiceHandler(application.NewSkillServiceServer(skillRepo), opts...)
	_, modelProviderHandler := agentsv1connect.NewModelProviderServiceHandler(modelProviderSvc, opts...)

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = os.Stderr
	gin.DefaultErrorWriter = os.Stderr
	r := gin.New()
	r.Use(gin.Recovery())
	cfg := &config.AppConfig{}
	cfg.HTTP.Greeting = "Hello, world!"
	dfx, err := newDaemonFixture(ctx, agentRepo)
	if err != nil {
		return nil, nil, err
	}
	r.Use(httpHandler.AuthMiddleware(cfg,
		func() auth.Repository { return authRepo },
		func() apitoken.Repository { return dfx.tokens },
		func() workspace.Repository { return wsRepo }))
	httpHandler.NewHealthHandler(service.NewHealthService(repo.NewHealthRepository(), cfg)).Register(r)
	// Keep unsupported methods inaccessible rather than simulating their behavior.
	for _, procedure := range []string{
		agentsv1connect.AuthServiceLoginProcedure,
		agentsv1connect.AuthServiceMeProcedure,
		agentsv1connect.AuthServiceLogoutProcedure,
		agentsv1connect.AuthServiceListOAuthProvidersProcedure,
		agentsv1connect.AuthServiceUpdateProfileProcedure,
		agentsv1connect.AuthServiceChangePasswordProcedure,
		agentsv1connect.AuthServiceBeginOAuthFlowProcedure,
		agentsv1connect.AuthServiceCompleteOAuthFlowProcedure,
		agentsv1connect.AuthServiceRedeemOAuthHandoffProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", authHandler)))
	}
	for _, procedure := range []string{
		agentsv1connect.WorkspaceServiceListWorkspacesProcedure,
		agentsv1connect.WorkspaceServiceGetWorkspaceProcedure,
		agentsv1connect.WorkspaceServiceCreateWorkspaceProcedure,
		agentsv1connect.WorkspaceServiceUpdateWorkspaceProcedure,
		agentsv1connect.WorkspaceServiceDeleteWorkspaceProcedure,
		agentsv1connect.WorkspaceServiceListWorkspaceMembersProcedure,
		agentsv1connect.WorkspaceServiceAddWorkspaceMemberProcedure,
		agentsv1connect.WorkspaceServiceUpdateWorkspaceMemberProcedure,
		agentsv1connect.WorkspaceServiceRemoveWorkspaceMemberProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", wsHandler)))
	}
	// Exercise the real REST handler's honest storage-not-configured response.
	httpHandler.NewUploadHandler(nil).Register(r)
	for _, procedure := range []string{
		agentsv1connect.AgentServiceListAgentsProcedure,
		agentsv1connect.AgentServiceGetAgentProcedure,
		agentsv1connect.AgentServiceCreateAgentProcedure,
		agentsv1connect.AgentServiceUpdateAgentProcedure,
		agentsv1connect.AgentServiceUpdateAgentConfigurationProcedure,
		agentsv1connect.AgentServiceDeleteAgentProcedure,
		agentsv1connect.AgentServiceRestoreAgentProcedure,
		agentsv1connect.AgentServiceGetAgentOperationProcedure,
		agentsv1connect.AgentServiceListAgentOperationsProcedure,
		agentsv1connect.AgentServiceRetryAgentOperationProcedure,
		agentsv1connect.AgentServiceGetAgentRuntimeStatusProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", agentHandler)))
	}
	// Skill selection is read-only; no standalone CRUD or content/resource API.
	r.POST("/api"+agentsv1connect.SkillServiceListSkillsProcedure,
		gin.WrapH(http.StripPrefix("/api", skillHandler)))
	repoFx.register(r, opts)
	for _, procedure := range []string{
		agentsv1connect.ModelProviderServiceListModelProvidersProcedure,
		agentsv1connect.ModelProviderServiceGetModelProviderProcedure,
		agentsv1connect.ModelProviderServiceCreateModelProviderProcedure,
		agentsv1connect.ModelProviderServiceUpdateModelProviderProcedure,
		agentsv1connect.ModelProviderServiceDeleteModelProviderProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", modelProviderHandler)))
	}
	backends := newFileBackends()
	backends.register(r, opts)
	boxRunner, chatSessions, runnerProviders, err := registerChat(ctx, r, agentRepo, wsRepo, opts, bx, backends, skillRepo)
	if err != nil {
		return nil, nil, err
	}
	rt := reloader{
		runner: boxRunner, agents: agentRepo,
		workspaces: []string{chatWorkspace, boxWorkspace, filesWorkspace, sessionsWorkspace, cronWorkspace},
		providers:  runnerProviders,
	}
	agentSvc.SetRuntime(rt)
	if err := registerCronFixture(ctx, r, opts, agentRepo, boxRunner, chatSessions, telegramRepo); err != nil {
		return nil, nil, err
	}
	if err := registerTelegramFixture(ctx, r, opts, wsRepo, agentRepo, telegramRepo); err != nil {
		return nil, nil, err
	}
	if err := rt.ReloadRunner(ctx); err != nil {
		return nil, nil, err
	}
	_, boxHandler := agentsv1connect.NewButterBoxServiceHandler(bx.svc, opts...)
	for _, procedure := range []string{
		agentsv1connect.ButterBoxServiceListButterBoxesProcedure,
		agentsv1connect.ButterBoxServiceGetButterBoxProcedure,
		agentsv1connect.ButterBoxServiceCreateButterBoxProcedure,
		agentsv1connect.ButterBoxServiceUpdateButterBoxProcedure,
		agentsv1connect.ButterBoxServiceDeleteButterBoxProcedure,
		agentsv1connect.ButterBoxServiceSetButterBoxTokenProcedure,
		agentsv1connect.ButterBoxServiceGetButterBoxStatusProcedure,
		agentsv1connect.ButterBoxServiceListButterBoxModelsProcedure,
		agentsv1connect.ButterBoxServiceListButterBoxCursorModelsProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", boxHandler)))
	}
	memFx := newMemoryFixture(wsRepo)
	memFx.register(r, opts)
	mcpFx := newMCPFixture(agentRepo)
	mux := http.NewServeMux()
	mcpFx.register(r, mux, opts)
	mcpFx.remoteSvc.SetDaemonRegistry(dfx.registry)
	dfx.register(r, mux, opts, wsRepo)
	mux.Handle("/fake-box/", boxDouble())
	mux.Handle("/butter/", http.StripPrefix("/butter", r))
	mux.Handle(legacyPrefix+"/", legacyMCPServer(r))
	mux.Handle("/", r)
	return mcpFx.wrap(mux), dfx, nil
}

func run() error {
	addr := flag.String("addr", "127.0.0.1:0", "loopback listen address (port 0 selects a free port)")
	flag.Parse()
	host, _, err := net.SplitHostPort(*addr)
	if err != nil {
		return err
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		return fmt.Errorf("development fixture requires a loopback address: %q", *addr)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	handler, err := fixture(ctx)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	// stdout is a machine-readable readiness channel; all logs go to stderr.
	if _, err := fmt.Fprintln(os.Stdout, "http://"+listener.Addr().String()); err != nil {
		return err
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	select {
	case err := <-served:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err := run(); err != nil {
		slog.Error("contract fixture failed", "err", err)
		os.Exit(1)
	}
}
