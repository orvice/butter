package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/auth"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const sessionSecretLen = 32

type AuthServiceServer struct {
	repo       auth.Repository
	wsRepo     workspacerepo.Repository
	sessionTTL time.Duration
}

func NewAuthServiceServer(repo auth.Repository, sessionTTL time.Duration) *AuthServiceServer {
	if sessionTTL <= 0 {
		sessionTTL = 7 * 24 * time.Hour
	}
	return &AuthServiceServer{repo: repo, sessionTTL: sessionTTL}
}

func (s *AuthServiceServer) SetRepo(repo auth.Repository) {
	s.repo = repo
}

// SetWorkspaceRepo wires the workspace repository so Login responses can
// carry the user's workspace memberships.
func (s *AuthServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.wsRepo = repo
}

func (s *AuthServiceServer) Login(ctx context.Context, req *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "auth store not available")
	}
	username := strings.TrimSpace(req.GetUsername())
	if username == "" {
		return nil, twirp.RequiredArgumentError("username")
	}
	if req.GetPassword() == "" {
		return nil, twirp.RequiredArgumentError("password")
	}

	user, passwordHash, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NewError(twirp.Unauthenticated, "invalid username or password")
		}
		return nil, twirp.InternalErrorWith(err)
	}
	if user.GetDisabled() {
		return nil, twirp.NewError(twirp.PermissionDenied, "user disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.GetPassword())); err != nil {
		return nil, twirp.NewError(twirp.Unauthenticated, "invalid username or password")
	}

	secret, err := generateSessionSecret()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	now := time.Now().UTC()
	expiresAt := now.Add(s.sessionTTL)
	session := &auth.Session{
		ID:        uuid.NewString(),
		UserID:    user.GetId(),
		TokenHash: HashAuthSessionToken(secret),
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	workspaces := s.userWorkspaces(ctx, user)
	return &agentsv1.LoginResponse{
		Token:      secret,
		User:       user,
		ExpiresAt:  timestamppb.New(expiresAt),
		Workspaces: workspaces,
	}, nil
}

// userWorkspaces returns the workspaces the user can access. Global admins
// (role == "admin") receive every workspace; other users receive only their
// memberships. Returns nil on any error so login is not blocked by lookup
// failures.
func (s *AuthServiceServer) userWorkspaces(ctx context.Context, user *agentsv1.User) []*agentsv1.Workspace {
	if s.wsRepo == nil || user == nil {
		return nil
	}
	if user.GetRole() == "admin" {
		all, err := s.wsRepo.ListWorkspaces(ctx)
		if err != nil {
			return nil
		}
		return all
	}
	members, err := s.wsRepo.ListMembershipsForUser(ctx, user.GetId())
	if err != nil {
		return nil
	}
	out := make([]*agentsv1.Workspace, 0, len(members))
	for _, m := range members {
		ws, err := s.wsRepo.GetWorkspace(ctx, m.GetWorkspaceId())
		if err != nil {
			continue
		}
		out = append(out, ws)
	}
	return out
}

// BootstrapDefaultWorkspace ensures a "default" workspace exists when the
// workspace store is empty. The initial admin (if present) is added as the
// initial owner so the dashboard always has at least one workspace to enter.
func BootstrapDefaultWorkspace(ctx context.Context, wsRepo workspacerepo.Repository, authRepo auth.Repository) error {
	if wsRepo == nil {
		return nil
	}
	count, err := wsRepo.CountWorkspaces(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now().UTC()
	ws := &agentsv1.Workspace{
		Id:        uuid.NewString(),
		Name:      "Default",
		Slug:      wsctx.DefaultSlug,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}
	created, err := wsRepo.CreateWorkspace(ctx, ws)
	if err != nil {
		return err
	}
	if authRepo == nil {
		return nil
	}
	users, err := authRepo.ListUsers(ctx)
	if err != nil || len(users) == 0 {
		return nil
	}
	for _, u := range users {
		_, _ = wsRepo.AddMember(ctx, &agentsv1.WorkspaceMember{
			WorkspaceId: created.GetId(),
			UserId:      u.GetId(),
			Role:        "owner",
			CreatedAt:   timestamppb.New(now),
		})
	}
	return nil
}

func (s *AuthServiceServer) Me(ctx context.Context, _ *agentsv1.MeRequest) (*agentsv1.MeResponse, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, twirp.NewError(twirp.Unauthenticated, "unauthenticated")
	}
	return &agentsv1.MeResponse{User: user}, nil
}

func (s *AuthServiceServer) Logout(ctx context.Context, _ *agentsv1.LogoutRequest) (*agentsv1.LogoutResponse, error) {
	if s.repo == nil {
		return &agentsv1.LogoutResponse{}, nil
	}
	session, ok := auth.SessionFromContext(ctx)
	if !ok {
		return nil, twirp.NewError(twirp.Unauthenticated, "unauthenticated")
	}
	if err := s.repo.RevokeSession(ctx, session.ID); err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.LogoutResponse{}, nil
}

func (s *AuthServiceServer) ListUsers(ctx context.Context, _ *agentsv1.ListUsersRequest) (*agentsv1.ListUsersResponse, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListUsersResponse{Users: users}, nil
}

func (s *AuthServiceServer) CreateUser(ctx context.Context, req *agentsv1.CreateUserRequest) (*agentsv1.CreateUserResponse, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	username := strings.TrimSpace(req.GetUsername())
	if username == "" {
		return nil, twirp.RequiredArgumentError("username")
	}
	if req.GetPassword() == "" {
		return nil, twirp.RequiredArgumentError("password")
	}
	role := normalizeRole(req.GetRole())
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	now := time.Now().UTC()
	user := &agentsv1.User{
		Id:          uuid.NewString(),
		Username:    username,
		DisplayName: strings.TrimSpace(req.GetDisplayName()),
		Role:        role,
		Disabled:    req.GetDisabled(),
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}
	if user.GetDisplayName() == "" {
		user.DisplayName = username
	}
	if err := s.repo.CreateUser(ctx, user, string(hash)); err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, twirp.NewError(twirp.AlreadyExists, "username already exists")
		}
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.CreateUserResponse{User: user}, nil
}

func (s *AuthServiceServer) UpdateUserPassword(ctx context.Context, req *agentsv1.UpdateUserPasswordRequest) (*agentsv1.UpdateUserPasswordResponse, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.GetId())
	if id == "" {
		return nil, twirp.RequiredArgumentError("id")
	}
	if req.GetPassword() == "" {
		return nil, twirp.RequiredArgumentError("password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	user, err := s.repo.UpdateUserPassword(ctx, id, string(hash), time.Now().UTC())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NotFoundError("user")
		}
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.UpdateUserPasswordResponse{User: user}, nil
}

func (s *AuthServiceServer) SetUserDisabled(ctx context.Context, req *agentsv1.SetUserDisabledRequest) (*agentsv1.SetUserDisabledResponse, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.GetId())
	if id == "" {
		return nil, twirp.RequiredArgumentError("id")
	}
	current, _ := auth.UserFromContext(ctx)
	if current != nil && current.GetId() == id && req.GetDisabled() {
		return nil, twirp.NewError(twirp.PermissionDenied, "cannot disable current user")
	}
	user, err := s.repo.SetUserDisabled(ctx, id, req.GetDisabled(), time.Now().UTC())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NotFoundError("user")
		}
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.SetUserDisabledResponse{User: user}, nil
}

func (s *AuthServiceServer) requireAdmin(ctx context.Context) error {
	if s.repo == nil {
		return twirp.NewError(twirp.FailedPrecondition, "auth store not available")
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return twirp.NewError(twirp.Unauthenticated, "unauthenticated")
	}
	if user.GetRole() != "admin" {
		return twirp.NewError(twirp.PermissionDenied, "admin role required")
	}
	return nil
}

func normalizeRole(role string) string {
	role = strings.TrimSpace(strings.ToLower(role))
	if role == "" {
		return "user"
	}
	return role
}

func BootstrapInitialAdmin(ctx context.Context, repo auth.Repository, cfg config.AuthConfig) error {
	if repo == nil {
		return nil
	}
	if err := repo.EnsureIndexes(ctx); err != nil {
		return err
	}
	count, err := repo.CountUsers(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	username := strings.TrimSpace(cfg.InitialAdminUsername)
	password := cfg.InitialAdminPassword
	if username == "" || password == "" {
		return errors.New("auth initial admin is required: set auth.initial_admin_username and auth.initial_admin_password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	user := &agentsv1.User{
		Id:          uuid.NewString(),
		Username:    username,
		DisplayName: username,
		Role:        "admin",
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}
	return repo.CreateUser(ctx, user, string(hash))
}

func HashAuthSessionToken(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func generateSessionSecret() (string, error) {
	buf := make([]byte, sessionSecretLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "bs_" + hex.EncodeToString(buf), nil
}
