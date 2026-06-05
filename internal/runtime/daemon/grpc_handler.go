package daemon

import (
	"crypto/sha256"
	"encoding/hex"
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

	registry   *Registry
	tokenRepo  apitoken.Repository
	daemonRepo configrepo.DaemonConfigRepository
}

// NewGRPCHandler creates a new handler for daemon connections.
func NewGRPCHandler(registry *Registry, tokenRepo apitoken.Repository, daemonRepo configrepo.DaemonConfigRepository) *GRPCHandler {
	return &GRPCHandler{
		registry:   registry,
		tokenRepo:  tokenRepo,
		daemonRepo: daemonRepo,
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
	if regInfo.DaemonId == "" {
		return status.Errorf(codes.InvalidArgument, "daemon_id is required")
	}
	authInfo, err := h.authenticate(stream, regInfo)
	if err != nil {
		return err
	}
	regInfo.WorkspaceId = authInfo.workspaceID
	regInfo.Capabilities = filterAllowedCapabilities(regInfo.GetCapabilities(), authInfo.allowedCapabilities)
	if len(regInfo.GetCapabilities()) == 0 {
		return status.Errorf(codes.PermissionDenied, "daemon has no allowed capabilities")
	}

	conn := NewConnection(regInfo)
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		conn.RemoteAddr = p.Addr.String()
	}
	h.registry.Register(conn)
	defer func() {
		conn.Close()
		h.registry.Unregister(regInfo.GetWorkspaceId(), regInfo.DaemonId)
		logger.Info("daemon disconnected", "workspace_id", regInfo.GetWorkspaceId(), "daemon_id", regInfo.DaemonId)
	}()

	logger.Info("daemon connected",
		"workspace_id", regInfo.GetWorkspaceId(),
		"daemon_id", regInfo.DaemonId,
		"name", regInfo.Name,
		"capabilities", regInfo.Capabilities,
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
		logger.Debug("daemon stream error", "daemon_id", regInfo.DaemonId, "err", err)
	case <-ctx.Done():
	}

	wg.Wait()
	return nil
}

type authResult struct {
	workspaceID         string
	allowedCapabilities []string
}

func (h *GRPCHandler) authenticate(stream agentsv1.DaemonConnectorService_ConnectServer, regInfo *agentsv1.DaemonInfo) (*authResult, error) {
	if h.tokenRepo == nil || h.daemonRepo == nil {
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
		return nil, status.Errorf(codes.PermissionDenied, "token is not a daemon credential")
	}
	if stored.GetWorkspaceId() == "" {
		return nil, status.Errorf(codes.PermissionDenied, "daemon credential has no workspace")
	}
	if stored.GetDaemonId() != regInfo.GetDaemonId() {
		return nil, status.Errorf(codes.PermissionDenied, "daemon credential does not match daemon_id")
	}
	if expires := stored.GetExpiresAt(); expires != nil && time.Now().UTC().After(expires.AsTime()) {
		return nil, status.Errorf(codes.Unauthenticated, "daemon credential expired")
	}
	if !hasScope(stored.GetScopes(), "daemon:connect") {
		return nil, status.Errorf(codes.PermissionDenied, "daemon credential lacks daemon:connect scope")
	}
	if regInfo.GetWorkspaceId() != "" && regInfo.GetWorkspaceId() != stored.GetWorkspaceId() {
		return nil, status.Errorf(codes.PermissionDenied, "workspace_id does not match daemon credential")
	}

	cfg, err := h.daemonRepo.GetDaemonConfig(stream.Context(), stored.GetWorkspaceId(), stored.GetDaemonId())
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "daemon is not registered in workspace")
	}

	_ = h.tokenRepo.TouchLastUsed(stream.Context(), stored.GetId())
	return &authResult{workspaceID: stored.GetWorkspaceId(), allowedCapabilities: cfg.GetAllowedCapabilities()}, nil
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

func filterAllowedCapabilities(offered, allowed []string) []string {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, cap := range allowed {
		if cap != "" {
			allowedSet[cap] = struct{}{}
		}
	}
	var out []string
	for _, cap := range offered {
		if _, ok := allowedSet[cap]; ok {
			out = append(out, cap)
		}
	}
	return out
}
