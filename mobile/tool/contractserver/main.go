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

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/auth/provider"
	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/repo/auth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	oauthhandoffmemory "go.orx.me/apps/butter/internal/repo/oauthhandoff/memory"
	oauthstatememory "go.orx.me/apps/butter/internal/repo/oauthstate/memory"
	"go.orx.me/apps/butter/internal/repo/workspace"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/service"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// Only the auth boundary needed by Login and AuthMiddleware is implemented.
// Unsupported RPCs are not registered, so the embedded interface is never used.
// Session copies keep the middleware's asynchronous TouchSession race-free.
type authStore struct {
	auth.Repository
	mu       sync.Mutex
	user     *agentsv1.User
	hash     string
	sessions map[string]auth.Session
}

func (s *authStore) FindUserByUsername(_ context.Context, username string) (*agentsv1.User, string, error) {
	if username != s.user.GetUsername() {
		return nil, "", auth.ErrUserNotFound
	}
	return proto.Clone(s.user).(*agentsv1.User), s.hash, nil
}

func (s *authStore) GetUser(_ context.Context, id string) (*agentsv1.User, error) {
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
	// These public credentials must never be used outside this loopback fixture.
	hash, err := bcrypt.GenerateFromPassword([]byte("contract-test-password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	authRepo := &authStore{
		user: &agentsv1.User{Id: "mobile-test", Username: "mobile-test", DisplayName: "Mobile Test", Role: "user"},
		hash: string(hash), sessions: make(map[string]auth.Session),
	}
	wsRepo := workspacememory.New()
	agentRepo := configmemory.New()
	for _, id := range []string{"alpha", "beta"} {
		if _, err := wsRepo.CreateWorkspace(ctx, &agentsv1.Workspace{Id: id, Name: id, Slug: id}); err != nil {
			return nil, err
		}
		if _, err := wsRepo.AddMember(ctx, &agentsv1.WorkspaceMember{WorkspaceId: id, UserId: authRepo.user.GetId(), Role: "member"}); err != nil {
			return nil, err
		}
		if _, err := agentRepo.CreateAgent(ctx, id, &agentsv1.Agent{AgentId: "shared", Name: id + "-agent"}); err != nil {
			return nil, err
		}
	}
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
	_, agentHandler := agentsv1connect.NewAgentServiceHandler(application.NewAgentServiceServer(agentRepo), opts...)

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = os.Stderr
	gin.DefaultErrorWriter = os.Stderr
	r := gin.New()
	r.Use(gin.Recovery())
	cfg := &config.AppConfig{}
	cfg.HTTP.Greeting = "Hello, world!"
	r.Use(httpHandler.AuthMiddleware(cfg,
		func() auth.Repository { return authRepo }, nil,
		func() workspace.Repository { return wsRepo }))
	httpHandler.NewHealthHandler(service.NewHealthService(repo.NewHealthRepository(), cfg)).Register(r)
	// Keep unsupported methods inaccessible rather than simulating their behavior.
	for _, procedure := range []string{
		agentsv1connect.AuthServiceLoginProcedure,
		agentsv1connect.AuthServiceMeProcedure,
		agentsv1connect.AuthServiceLogoutProcedure,
		agentsv1connect.AuthServiceListOAuthProvidersProcedure,
		agentsv1connect.AuthServiceBeginOAuthFlowProcedure,
		agentsv1connect.AuthServiceCompleteOAuthFlowProcedure,
		agentsv1connect.AuthServiceRedeemOAuthHandoffProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", authHandler)))
	}
	r.POST("/api"+agentsv1connect.WorkspaceServiceListWorkspacesProcedure, gin.WrapH(http.StripPrefix("/api", wsHandler)))
	r.POST("/api"+agentsv1connect.AgentServiceListAgentsProcedure, gin.WrapH(http.StripPrefix("/api", agentHandler)))
	mux := http.NewServeMux()
	mux.Handle("/butter/", http.StripPrefix("/butter", r))
	mux.Handle("/", r)
	return mux, nil
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
