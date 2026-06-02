package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/apitoken"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// tokenSecretLen is the number of random bytes used in each token (excluding
// the "bt_" prefix). 24 bytes → 48 hex chars → 51 char full secret.
const tokenSecretLen = 24

// tokenPrefixLen is the visible portion of the secret stored as `prefix`.
const tokenPrefixLen = 12

// APITokenServiceServer manages API bearer tokens.
type APITokenServiceServer struct {
	repo apitoken.Repository
}

func NewAPITokenServiceServer(repo apitoken.Repository) *APITokenServiceServer {
	return &APITokenServiceServer{repo: repo}
}

// SetRepo swaps the underlying repository. Used to attach the persistent
// backend after bootstrap.
func (s *APITokenServiceServer) SetRepo(repo apitoken.Repository) {
	s.repo = repo
}

func (s *APITokenServiceServer) ListAPITokens(ctx context.Context, _ *agentsv1.ListAPITokensRequest) (*agentsv1.ListAPITokensResponse, error) {
	if s.repo == nil {
		return &agentsv1.ListAPITokensResponse{}, nil
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	tokens, err := s.repo.List(ctx, wsID)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.ListAPITokensResponse{Tokens: tokens}, nil
}

func (s *APITokenServiceServer) CreateAPIToken(ctx context.Context, req *agentsv1.CreateAPITokenRequest) (*agentsv1.CreateAPITokenResponse, error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("api token store not available"))
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	name := req.GetName()
	if name == "" {
		return nil, connectx.RequiredArgument("name")
	}

	secret, err := generateSecret()
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	hash := HashAPITokenSecret(secret)

	token := &agentsv1.APIToken{
		Id:          uuid.NewString(),
		Name:        name,
		Prefix:      tokenPrefix(secret),
		CreatedAt:   timestamppb.New(time.Now().UTC()),
		WorkspaceId: wsID,
	}
	logger := log.FromContext(ctx)
	if err := s.repo.Create(ctx, token, hash); err != nil {
		logger.Error("create api token failed", "workspace_id", wsID, "name", name, "err", err)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("api token created", "workspace_id", wsID, "id", token.GetId(), "name", name, "prefix", token.GetPrefix())
	return &agentsv1.CreateAPITokenResponse{Token: token, Secret: secret}, nil
}

func (s *APITokenServiceServer) RevokeAPIToken(ctx context.Context, req *agentsv1.RevokeAPITokenRequest) (*agentsv1.RevokeAPITokenResponse, error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("api token store not available"))
	}
	if req.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, apitoken.ErrNotFound) {
			return nil, connectx.NotFound("api token not found")
		}
		return nil, connectx.InternalWith(err)
	}
	// Disguise cross-tenant revoke attempts as NotFound to avoid leaking the
	// existence of tokens belonging to other workspaces.
	if existing.GetWorkspaceId() != wsID {
		return nil, connectx.NotFound("api token not found")
	}
	logger := log.FromContext(ctx)
	token, err := s.repo.Revoke(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, apitoken.ErrNotFound) {
			return nil, connectx.NotFound("api token not found")
		}
		logger.Error("revoke api token failed", "workspace_id", wsID, "id", req.GetId(), "err", err)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("api token revoked", "workspace_id", wsID, "id", token.GetId(), "name", token.GetName())
	return &agentsv1.RevokeAPITokenResponse{Token: token}, nil
}

// HashAPITokenSecret is shared between the service and the auth middleware so
// they agree on the hash function for stored tokens.
func HashAPITokenSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func generateSecret() (string, error) {
	buf := make([]byte, tokenSecretLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "bt_" + hex.EncodeToString(buf), nil
}

func tokenPrefix(secret string) string {
	if len(secret) <= tokenPrefixLen {
		return secret
	}
	return secret[:tokenPrefixLen]
}
