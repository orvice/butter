package daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/apitoken"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

var daemonHeartbeatInterval = 20 * time.Second

// GRPCHandler implements the DaemonConnectorService over ConnectRPC.
type GRPCHandler struct {
	agentsv1connect.UnimplementedDaemonConnectorServiceHandler

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

// SetAPITokenRepo swaps the repository used to authenticate daemon runtime
// tokens. Routes are registered before bootstrap wires persistent storage, so
// the app layer calls this once the token repository is available.
func (h *GRPCHandler) SetAPITokenRepo(repo apitoken.Repository) {
	h.tokenRepo = repo
}

// Connect implements the bidirectional streaming RPC.
func (h *GRPCHandler) Connect(ctx context.Context, stream *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse]) error {
	logger := log.FromContext(ctx)

	// First message must be a register.
	firstMsg, err := stream.Receive()
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to receive first message: %w", err))
	}
	regInfo := firstMsg.GetRegister()
	if regInfo == nil {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("first message must be a register"))
	}
	authInfo, err := h.authenticate(ctx, stream, regInfo)
	if err != nil {
		return err
	}
	regInfo.WorkspaceId = authInfo.workspaceID
	regInfo.DaemonRuntimeId = authInfo.runtimeID
	if len(regInfo.GetAcpRuntimes()) == 0 {
		regInfo.AcpRuntimes = []string{"opencode", "codex"}
	}

	conn := NewConnection(regInfo)
	if peer := stream.Peer(); peer.Addr != "" {
		conn.RemoteAddr = peer.Addr
	}
	if err := h.registry.Register(conn); err != nil {
		if errors.Is(err, ErrRuntimeAlreadyConnected) {
			return connect.NewError(connect.CodeAlreadyExists, errors.New("daemon runtime already connected"))
		}
		return connect.NewError(connect.CodeInternal, fmt.Errorf("register daemon runtime: %w", err))
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
	done := make(chan struct{})

	// Send loop: conn.SendCh → stream.Send().
	wg.Add(1)
	go func() {
		defer wg.Done()
		heartbeat := time.NewTicker(daemonHeartbeatInterval)
		defer heartbeat.Stop()
		for {
			select {
			case msg := <-conn.SendCh:
				if err := stream.Send(msg); err != nil {
					errCh <- fmt.Errorf("send: %w", err)
					return
				}
			case <-heartbeat.C:
				if err := stream.Send(&agentsv1.ConnectResponse{}); err != nil {
					errCh <- fmt.Errorf("heartbeat: %w", err)
					return
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Recv loop: stream.Receive() → conn.DispatchUpdate().
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := stream.Receive()
			if err != nil {
				if errors.Is(err, io.EOF) {
					errCh <- nil
				} else {
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
		if err != nil {
			logger.Debug("daemon stream error", "daemon_runtime_id", regInfo.GetDaemonRuntimeId(), "err", err)
		}
	case <-ctx.Done():
	}

	close(done)
	wg.Wait()
	return nil
}

type authResult struct {
	workspaceID string
	runtimeID   string
}

func (h *GRPCHandler) authenticate(ctx context.Context, stream *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse], regInfo *agentsv1.DaemonInfo) (*authResult, error) {
	if h.tokenRepo == nil || h.runtimeRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("daemon auth repositories not wired"))
	}

	authz := stream.RequestHeader().Get("Authorization")
	if authz == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authorization header"))
	}

	token := strings.TrimPrefix(authz, "Bearer ")
	stored, err := h.tokenRepo.Lookup(ctx, hashSecret(token))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
	}
	if stored.GetKind() != agentsv1.APITokenKind_API_TOKEN_KIND_DAEMON {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("token is not a daemon runtime token"))
	}
	if stored.GetWorkspaceId() == "" {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("daemon runtime token has no workspace"))
	}
	if stored.GetDaemonRuntimeId() == "" {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("daemon runtime token has no runtime"))
	}
	if regInfo.GetDaemonRuntimeId() != "" && stored.GetDaemonRuntimeId() != regInfo.GetDaemonRuntimeId() {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("daemon_runtime_id does not match daemon runtime token"))
	}
	if expires := stored.GetExpiresAt(); expires != nil && time.Now().UTC().After(expires.AsTime()) {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("daemon runtime token expired"))
	}
	if !hasScope(stored.GetScopes(), "daemon:connect") {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("daemon runtime token lacks daemon:connect scope"))
	}
	if regInfo.GetWorkspaceId() != "" && regInfo.GetWorkspaceId() != stored.GetWorkspaceId() {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("workspace_id does not match daemon runtime token"))
	}

	if _, err := h.runtimeRepo.GetDaemonRuntime(ctx, stored.GetWorkspaceId(), stored.GetDaemonRuntimeId()); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("daemon runtime is not registered in workspace"))
	}

	_ = h.tokenRepo.TouchLastUsed(ctx, stored.GetId())
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
