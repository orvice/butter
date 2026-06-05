package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"go.orx.me/apps/butter/internal/repo/apitoken"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// GRPCHandler implements the DaemonConnectorService gRPC service.
type GRPCHandler struct {
	agentsv1.UnimplementedDaemonConnectorServiceServer

	registry    *Registry
	tokenRepo   apitoken.Repository
	runtimeRepo configrepo.DaemonRuntimeRepository
}

// NewGRPCHandler creates a new handler for daemon connections.
func NewGRPCHandler(registry *Registry, tokenRepo apitoken.Repository, runtimeRepo configrepo.DaemonRuntimeRepository) *GRPCHandler {
	return &GRPCHandler{
		registry:    registry,
		tokenRepo:   tokenRepo,
		runtimeRepo: runtimeRepo,
	}
}

// Connect implements the bidirectional streaming RPC.
func (h *GRPCHandler) Connect(stream agentsv1.DaemonConnectorService_ConnectServer) error {
	ctx := stream.Context()
	logger := log.FromContext(ctx)

	// First message must be a register.
	firstMsg, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to receive first message: %v", err)
	}
	regInfo := firstMsg.GetRegister()
	if regInfo == nil {
		return status.Errorf(codes.InvalidArgument, "first message must be a register")
	}
	authInfo, err := h.authenticate(stream, regInfo)
	if err != nil {
		return err
	}
	regInfo.WorkspaceId = authInfo.workspaceID
	regInfo.DaemonRuntimeId = authInfo.runtimeID
	if len(regInfo.GetAcpRuntimes()) == 0 {
		regInfo.AcpRuntimes = []string{"opencode", "codex"}
	}

	conn := NewConnection(regInfo)
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		conn.RemoteAddr = p.Addr.String()
	}
	if err := h.registry.Register(conn); err != nil {
		if errors.Is(err, ErrRuntimeAlreadyConnected) {
			return status.Errorf(codes.AlreadyExists, "daemon runtime already connected")
		}
		return status.Errorf(codes.Internal, "register daemon runtime: %v", err)
	}
	defer func() {
		conn.Close()
		h.registry.Unregister(regInfo.GetWorkspaceId(), regInfo.GetDaemonRuntimeId())
		logger.Info("daemon disconnected", "workspace_id", regInfo.GetWorkspaceId(), "daemon_runtime_id", regInfo.GetDaemonRuntimeId())
	}()

	logger.Info("daemon connected",
		"workspace_id", regInfo.GetWorkspaceId(),
		"daemon_runtime_id", regInfo.GetDaemonRuntimeId(),
		"name", regInfo.Name,
		"acp_runtimes", regInfo.AcpRuntimes,
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// Send loop: conn.SendCh → stream.Send().
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case msg := <-conn.SendCh:
				if err := stream.Send(msg); err != nil {
					errCh <- fmt.Errorf("send: %w", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Recv loop: stream.Recv() → conn.DispatchUpdate().
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					errCh <- fmt.Errorf("recv: %w", err)
				}
				return
			}
			if update := msg.GetTaskUpdate(); update != nil {
				conn.DispatchUpdate(update)
			}
		}
	}()

	// Wait for either loop to finish or context to cancel.
	select {
	case err := <-errCh:
		logger.Debug("daemon stream error", "daemon_runtime_id", regInfo.GetDaemonRuntimeId(), "err", err)
	case <-ctx.Done():
	}

	wg.Wait()
	return nil
}

type authResult struct {
	workspaceID string
	runtimeID   string
}

func (h *GRPCHandler) authenticate(stream agentsv1.DaemonConnectorService_ConnectServer, regInfo *agentsv1.DaemonInfo) (*authResult, error) {
	if h.tokenRepo == nil || h.runtimeRepo == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "daemon auth repositories not wired")
	}

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")
	stored, err := h.tokenRepo.Lookup(stream.Context(), hashSecret(token))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	if stored.GetKind() != agentsv1.APITokenKind_API_TOKEN_KIND_DAEMON {
		return nil, status.Errorf(codes.PermissionDenied, "token is not a daemon runtime token")
	}
	if stored.GetWorkspaceId() == "" {
		return nil, status.Errorf(codes.PermissionDenied, "daemon runtime token has no workspace")
	}
	if stored.GetDaemonRuntimeId() == "" {
		return nil, status.Errorf(codes.PermissionDenied, "daemon runtime token has no runtime")
	}
	if regInfo.GetDaemonRuntimeId() != "" && stored.GetDaemonRuntimeId() != regInfo.GetDaemonRuntimeId() {
		return nil, status.Errorf(codes.PermissionDenied, "daemon_runtime_id does not match daemon runtime token")
	}
	if expires := stored.GetExpiresAt(); expires != nil && time.Now().UTC().After(expires.AsTime()) {
		return nil, status.Errorf(codes.Unauthenticated, "daemon runtime token expired")
	}
	if !hasScope(stored.GetScopes(), "daemon:connect") {
		return nil, status.Errorf(codes.PermissionDenied, "daemon runtime token lacks daemon:connect scope")
	}
	if regInfo.GetWorkspaceId() != "" && regInfo.GetWorkspaceId() != stored.GetWorkspaceId() {
		return nil, status.Errorf(codes.PermissionDenied, "workspace_id does not match daemon runtime token")
	}

	if _, err := h.runtimeRepo.GetDaemonRuntime(stream.Context(), stored.GetWorkspaceId(), stored.GetDaemonRuntimeId()); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "daemon runtime is not registered in workspace")
	}

	_ = h.tokenRepo.TouchLastUsed(stream.Context(), stored.GetId())
	return &authResult{workspaceID: stored.GetWorkspaceId(), runtimeID: stored.GetDaemonRuntimeId()}, nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func hasScope(scopes []string, want string) bool {
	for _, scope := range scopes {
		if scope == want || scope == "daemon:*" {
			return true
		}
	}
	return false
}
