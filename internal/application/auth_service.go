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
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const sessionSecretLen = 32

type AuthServiceServer struct {
	repo       auth.Repository
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

	return &agentsv1.LoginResponse{
		Token:     secret,
		User:      user,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
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
