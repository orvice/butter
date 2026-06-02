package daemon

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"butterfly.orx.me/core/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// GRPCHandler implements the DaemonConnectorService gRPC service.
type GRPCHandler struct {
	agentsv1.UnimplementedDaemonConnectorServiceServer

	registry         *Registry
	apiTokenProvider func() string
}

// NewGRPCHandler creates a new handler for daemon connections.
func NewGRPCHandler(registry *Registry, apiTokenProvider func() string) *GRPCHandler {
	return &GRPCHandler{
		registry:         registry,
		apiTokenProvider: apiTokenProvider,
	}
}

// Connect implements the bidirectional streaming RPC.
func (h *GRPCHandler) Connect(stream agentsv1.DaemonConnectorService_ConnectServer) error {
	ctx := stream.Context()
	logger := log.FromContext(ctx)

	// Authenticate.
	if err := h.authenticate(stream); err != nil {
		return err
	}

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

	conn := NewConnection(regInfo)
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		conn.RemoteAddr = p.Addr.String()
	}
	h.registry.Register(conn)
	defer func() {
		conn.Close()
		h.registry.Unregister(regInfo.DaemonId)
		logger.Info("daemon disconnected", "daemon_id", regInfo.DaemonId)
	}()

	logger.Info("daemon connected",
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

func (h *GRPCHandler) authenticate(stream agentsv1.DaemonConnectorService_ConnectServer) error {
	apiToken := strings.TrimSpace(h.apiToken())
	if apiToken == "" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")
	if token != apiToken {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}

	return nil
}

func (h *GRPCHandler) apiToken() string {
	if h.apiTokenProvider == nil {
		return ""
	}
	return h.apiTokenProvider()
}
