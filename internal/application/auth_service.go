package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
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
	logger := log.FromContext(ctx)
	if s.repo == nil {
		logger.Warn("login rejected: auth store not available")
		return nil, twirp.NewError(twirp.FailedPrecondition, "auth store not available")
	}
	username := strings.TrimSpace(req.GetUsername())
	if username == "" {
		return nil, twirp.RequiredArgumentError("username")
	}
	if req.GetPassword() == "" {
		return nil, twirp.RequiredArgumentError("password")
	}

	logger.Debug("login attempt", "username", username)

	user, passwordHash, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			logger.Info("login failed: unknown user", "username", username)
			return nil, twirp.NewError(twirp.Unauthenticated, "invalid username or password")
		}
		logger.Error("login failed: user lookup error", "username", username, "err", err)
		return nil, twirp.InternalErrorWith(err)
	}
	if user.GetDisabled() {
		logger.Info("login rejected: user disabled", "username", username, "user_id", user.GetId())
		return nil, twirp.NewError(twirp.PermissionDenied, "user disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.GetPassword())); err != nil {
		logger.Info("login failed: password mismatch", "username", username, "user_id", user.GetId())
		return nil, twirp.NewError(twirp.Unauthenticated, "invalid username or password")
	}

	secret, err := generateSessionSecret()
	if err != nil {
		logger.Error("login failed: cannot generate session secret", "user_id", user.GetId(), "err", err)
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
		logger.Error("login failed: cannot create session", "user_id", user.GetId(), "err", err)
		return nil, twirp.InternalErrorWith(err)
	}

	workspaces := s.userWorkspaces(ctx, user)
	logger.Info("login succeeded",
		"username", username,
		"user_id", user.GetId(),
		"role", user.GetRole(),
		"session_id", session.ID,
		"workspace_count", len(workspaces),
	)
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
	logger := log.FromContext(ctx)
	if s.repo == nil {
		return &agentsv1.LogoutResponse{}, nil
	}
	session, ok := auth.SessionFromContext(ctx)
	if !ok {
		return nil, twirp.NewError(twirp.Unauthenticated, "unauthenticated")
	}
	if err := s.repo.RevokeSession(ctx, session.ID); err != nil {
		logger.Error("logout failed: revoke session", "session_id", session.ID, "user_id", session.UserID, "err", err)
		return nil, twirp.InternalErrorWith(err)
	}
	logger.Info("logout succeeded", "session_id", session.ID, "user_id", session.UserID)
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
	logger := log.FromContext(ctx)
	if err := s.repo.CreateUser(ctx, user, string(hash)); err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			logger.Info("create user rejected: username already exists", "username", username)
			return nil, twirp.NewError(twirp.AlreadyExists, "username already exists")
		}
		logger.Error("create user failed", "username", username, "err", err)
		return nil, twirp.InternalErrorWith(err)
	}
	logger.Info("user created", "username", username, "user_id", user.GetId(), "role", user.GetRole(), "disabled", user.GetDisabled())
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
	logger := log.FromContext(ctx)
	user, err := s.repo.UpdateUserPassword(ctx, id, string(hash), time.Now().UTC())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NotFoundError("user")
		}
		logger.Error("update user password failed", "user_id", id, "err", err)
		return nil, twirp.InternalErrorWith(err)
	}
	logger.Info("user password updated", "user_id", id, "username", user.GetUsername())
	return &agentsv1.UpdateUserPasswordResponse{User: user}, nil
}

func (s *AuthServiceServer) UpdateProfile(ctx context.Context, req *agentsv1.UpdateProfileRequest) (*agentsv1.UpdateProfileResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "auth store not available")
	}
	current, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, twirp.NewError(twirp.Unauthenticated, "unauthenticated")
	}
	displayName := strings.TrimSpace(req.GetDisplayName())
	if displayName == "" {
		return nil, twirp.RequiredArgumentError("display_name")
	}
	avatarURL := strings.TrimSpace(req.GetAvatarUrl())
	user, err := s.repo.UpdateUserProfile(ctx, current.GetId(), displayName, avatarURL, time.Now().UTC())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NotFoundError("user")
		}
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.UpdateProfileResponse{User: user}, nil
}

func (s *AuthServiceServer) ChangePassword(ctx context.Context, req *agentsv1.ChangePasswordRequest) (*agentsv1.ChangePasswordResponse, error) {
	if s.repo == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "auth store not available")
	}
	current, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, twirp.NewError(twirp.Unauthenticated, "unauthenticated")
	}
	if req.GetCurrentPassword() == "" {
		return nil, twirp.RequiredArgumentError("current_password")
	}
	if req.GetNewPassword() == "" {
		return nil, twirp.RequiredArgumentError("new_password")
	}
	_, passwordHash, err := s.repo.FindUserByID(ctx, current.GetId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NotFoundError("user")
		}
		return nil, twirp.InternalErrorWith(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.GetCurrentPassword())); err != nil {
		return nil, twirp.NewError(twirp.PermissionDenied, "current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetNewPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	user, err := s.repo.UpdateUserPassword(ctx, current.GetId(), string(hash), time.Now().UTC())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NotFoundError("user")
		}
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ChangePasswordResponse{User: user}, nil
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
	logger := log.FromContext(ctx)
	user, err := s.repo.SetUserDisabled(ctx, id, req.GetDisabled(), time.Now().UTC())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, twirp.NotFoundError("user")
		}
		logger.Error("set user disabled failed", "user_id", id, "disabled", req.GetDisabled(), "err", err)
		return nil, twirp.InternalErrorWith(err)
	}
	logger.Info("user disabled flag updated", "user_id", id, "username", user.GetUsername(), "disabled", user.GetDisabled())
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
	logger := log.FromContext(ctx)
	if repo == nil {
		logger.Info("skipping initial admin bootstrap, auth repo is nil")
		return nil
	}

	username := strings.TrimSpace(cfg.InitialAdminUsername)
	password := cfg.InitialAdminPassword
	logger.Info("initial admin bootstrap started",
		"initial_admin_username_set", username != "",
		"initial_admin_password_set", password != "",
	)

	if err := repo.EnsureIndexes(ctx); err != nil {
		logger.Error("failed to ensure auth indexes", "err", err)
		return err
	}
	logger.Debug("auth indexes ensured")

	count, err := repo.CountUsers(ctx)
	if err != nil {
		logger.Error("failed to count auth users", "err", err)
		return err
	}
	logger.Info("auth user count loaded", "user_count", count)
	if count > 0 {
		logger.Info("skipping initial admin bootstrap, users already exist", "user_count", count)
		return nil
	}

	if username == "" || password == "" {
		logger.Error("initial admin config missing",
			"initial_admin_username_set", username != "",
			"initial_admin_password_set", password != "",
		)
		return errors.New("auth initial admin is required: set auth.initial_admin_username and auth.initial_admin_password")
	}

	logger.Info("creating initial admin user", "username", username, "role", "admin")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash initial admin password", "username", username, "err", err)
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
	if err := repo.CreateUser(ctx, user, string(hash)); err != nil {
		logger.Error("failed to create initial admin user", "username", username, "user_id", user.GetId(), "err", err)
		return err
	}
	logger.Info("initial admin user created", "username", username, "user_id", user.GetId(), "role", user.GetRole())
	return nil
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
