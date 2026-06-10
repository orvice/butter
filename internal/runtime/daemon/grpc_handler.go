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
	"google.golang.org/protobuf/types/known/emptypb"

	"go.orx.me/apps/butter/internal/repo/apitoken"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

var daemonHeartbeatInterval = 20 * time.Second
var daemonPollDefaultWait = 25 * time.Second
var daemonPollMaxWait = 30 * time.Second
var daemonPollIdleTimeout = 90 * time.Second

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
		heartbeatCount := 0
		for {
			select {
			case msg := <-conn.SendCh:
				if err := stream.Send(msg); err != nil {
					errCh <- fmt.Errorf("send: %w", err)
					return
				}
			case <-heartbeat.C:
				if err := stream.Send(&agentsv1.ConnectResponse{
					Message: &agentsv1.ConnectResponse_Heartbeat{Heartbeat: &emptypb.Empty{}},
				}); err != nil {
					errCh <- fmt.Errorf("heartbeat: %w", err)
					return
				}
				heartbeatCount++
				if heartbeatCount <= 3 {
					logger.Info("daemon heartbeat sent",
						"workspace_id", regInfo.GetWorkspaceId(),
						"daemon_runtime_id", regInfo.GetDaemonRuntimeId(),
						"count", heartbeatCount,
					)
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
		heartbeatCount := 0
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
			} else if msg.GetHeartbeat() != nil {
				heartbeatCount++
				if heartbeatCount <= 3 {
					logger.Info("daemon heartbeat received",
						"workspace_id", regInfo.GetWorkspaceId(),
						"daemon_runtime_id", regInfo.GetDaemonRuntimeId(),
						"count", heartbeatCount,
					)
				}
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

// Register announces a daemon runtime using unary long-poll transport.
func (h *GRPCHandler) Register(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServiceRegisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceRegisterResponse], error) {
	logger := log.FromContext(ctx)
	regInfo := req.Msg.GetDaemon()
	if regInfo == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("daemon is required"))
	}
	authInfo, err := h.authenticateDaemon(ctx, req.Header().Get("Authorization"), regInfo)
	if err != nil {
		return nil, err
	}
	regInfo.WorkspaceId = authInfo.workspaceID
	regInfo.DaemonRuntimeId = authInfo.runtimeID
	if len(regInfo.GetAcpRuntimes()) == 0 {
		regInfo.AcpRuntimes = []string{"opencode", "codex"}
	}

	conn := NewConnection(regInfo)
	conn.MarkPollMode()
	replaced := h.registry.RegisterOrReplace(conn)
	logger.Info("daemon registered",
		"workspace_id", regInfo.GetWorkspaceId(),
		"daemon_runtime_id", regInfo.GetDaemonRuntimeId(),
		"name", regInfo.GetName(),
		"acp_runtimes", regInfo.GetAcpRuntimes(),
		"transport", "poll",
		"replaced", replaced,
	)

	return connect.NewResponse(&agentsv1.DaemonConnectorServiceRegisterResponse{Daemon: regInfo}), nil
}

// Poll waits briefly for queued task/cancel messages for a daemon runtime.
func (h *GRPCHandler) Poll(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServicePollRequest]) (*connect.Response[agentsv1.DaemonConnectorServicePollResponse], error) {
	authInfo, err := h.authenticateDaemon(ctx, req.Header().Get("Authorization"), nil)
	if err != nil {
		return nil, err
	}
	h.registry.PruneStalePollConnections(daemonPollIdleTimeout)

	conn := h.registry.Get(authInfo.workspaceID, authInfo.runtimeID)
	if conn == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("daemon is not registered"))
	}
	conn.Touch()

	wait := daemonPollDefaultWait
	if req.Msg.GetWaitTimeout() != nil {
		wait = req.Msg.GetWaitTimeout().AsDuration()
	}
	if wait < 0 {
		wait = 0
	}
	if wait > daemonPollMaxWait {
		wait = daemonPollMaxWait
	}

	messages, err := pollMessages(ctx, conn, wait)
	if err != nil {
		return nil, err
	}
	conn.Touch()
	return connect.NewResponse(&agentsv1.DaemonConnectorServicePollResponse{Messages: messages}), nil
}

// ReportTaskUpdate routes a daemon task update to the waiting bridge.
func (h *GRPCHandler) ReportTaskUpdate(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServiceReportTaskUpdateRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceReportTaskUpdateResponse], error) {
	update := req.Msg.GetUpdate()
	if update == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("update is required"))
	}
	authInfo, err := h.authenticateDaemon(ctx, req.Header().Get("Authorization"), nil)
	if err != nil {
		return nil, err
	}
	conn := h.registry.Get(authInfo.workspaceID, authInfo.runtimeID)
	if conn == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("daemon is not registered"))
	}
	conn.Touch()
	conn.DispatchUpdate(update)
	return connect.NewResponse(&agentsv1.DaemonConnectorServiceReportTaskUpdateResponse{}), nil
}

// Unregister marks a poll-mode daemon runtime offline on graceful shutdown.
func (h *GRPCHandler) Unregister(ctx context.Context, req *connect.Request[agentsv1.DaemonConnectorServiceUnregisterRequest]) (*connect.Response[agentsv1.DaemonConnectorServiceUnregisterResponse], error) {
	logger := log.FromContext(ctx)
	authInfo, err := h.authenticateDaemon(ctx, req.Header().Get("Authorization"), nil)
	if err != nil {
		return nil, err
	}
	if conn := h.registry.Get(authInfo.workspaceID, authInfo.runtimeID); conn != nil {
		conn.Close()
		h.registry.Unregister(authInfo.workspaceID, authInfo.runtimeID)
		logger.Info("daemon unregistered", "workspace_id", authInfo.workspaceID, "daemon_runtime_id", authInfo.runtimeID, "transport", "poll")
	}
	return connect.NewResponse(&agentsv1.DaemonConnectorServiceUnregisterResponse{}), nil
}

func pollMessages(ctx context.Context, conn *Connection, wait time.Duration) ([]*agentsv1.ConnectResponse, error) {
	messages := make([]*agentsv1.ConnectResponse, 0, 4)
	if wait <= 0 {
		select {
		case msg := <-conn.SendCh:
			if msg != nil {
				messages = append(messages, msg)
			}
		default:
		}
		return drainMessages(conn, messages), nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case msg := <-conn.SendCh:
		if msg != nil {
			messages = append(messages, msg)
		}
	case <-timer.C:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return drainMessages(conn, messages), nil
}

func drainMessages(conn *Connection, messages []*agentsv1.ConnectResponse) []*agentsv1.ConnectResponse {
	for len(messages) < 8 {
		select {
		case msg := <-conn.SendCh:
			if msg != nil {
				messages = append(messages, msg)
			}
		default:
			return messages
		}
	}
	return messages
}

type authResult struct {
	workspaceID string
	runtimeID   string
}

func (h *GRPCHandler) authenticate(ctx context.Context, stream *connect.BidiStream[agentsv1.ConnectRequest, agentsv1.ConnectResponse], regInfo *agentsv1.DaemonInfo) (*authResult, error) {
	return h.authenticateDaemon(ctx, stream.RequestHeader().Get("Authorization"), regInfo)
}

func (h *GRPCHandler) authenticateDaemon(ctx context.Context, authz string, regInfo *agentsv1.DaemonInfo) (*authResult, error) {
	if h.tokenRepo == nil || h.runtimeRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("daemon auth repositories not wired"))
	}

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
	if regInfo != nil && regInfo.GetDaemonRuntimeId() != "" && stored.GetDaemonRuntimeId() != regInfo.GetDaemonRuntimeId() {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("daemon_runtime_id does not match daemon runtime token"))
	}
	if expires := stored.GetExpiresAt(); expires != nil && time.Now().UTC().After(expires.AsTime()) {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("daemon runtime token expired"))
	}
	if !hasScope(stored.GetScopes(), "daemon:connect") {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("daemon runtime token lacks daemon:connect scope"))
	}
	if regInfo != nil && regInfo.GetWorkspaceId() != "" && regInfo.GetWorkspaceId() != stored.GetWorkspaceId() {
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
