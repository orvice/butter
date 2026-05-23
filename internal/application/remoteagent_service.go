package application

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/twitchtv/twirp"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RemoteAgentServiceServer struct {
	repo      configrepo.RemoteAgentRepository
	runtime   ConfigRuntime
	daemonReg *daemon.Registry
}

func NewRemoteAgentServiceServer(repo configrepo.RemoteAgentRepository) *RemoteAgentServiceServer {
	return &RemoteAgentServiceServer{repo: repo}
}

func (s *RemoteAgentServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

func (s *RemoteAgentServiceServer) SetDaemonRegistry(reg *daemon.Registry) {
	s.daemonReg = reg
}

func (s *RemoteAgentServiceServer) ListRemoteAgents(ctx context.Context, _ *agentsv1.ListRemoteAgentsRequest) (*agentsv1.ListRemoteAgentsResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	agents, err := s.repo.ListRemoteAgents(ctx, wsID)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListRemoteAgentsResponse{RemoteAgents: agents}, nil
}

func (s *RemoteAgentServiceServer) GetRemoteAgent(ctx context.Context, req *agentsv1.GetRemoteAgentRequest) (*agentsv1.GetRemoteAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	r, err := s.repo.GetRemoteAgent(ctx, wsID, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) CreateRemoteAgent(ctx context.Context, req *agentsv1.CreateRemoteAgentRequest) (*agentsv1.CreateRemoteAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateRemoteAgentURL(req.GetRemoteAgent()); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating remote agent",
		"workspace_id", wsID,
		"name", req.GetRemoteAgent().GetName(),
		"protocol", req.GetRemoteAgent().GetProtocol().String(),
	)
	r, err := mutateWithRuntime(
		func() (*agentsv1.RemoteAgent, error) {
			return s.repo.CreateRemoteAgent(ctx, wsID, req.GetRemoteAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteRemoteAgent(ctx, wsID, req.GetRemoteAgent().GetId()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("create remote agent failed", "workspace_id", wsID, "name", req.GetRemoteAgent().GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("remote agent created", "workspace_id", wsID, "id", r.GetId(), "name", r.GetName())
	return &agentsv1.CreateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) UpdateRemoteAgent(ctx context.Context, req *agentsv1.UpdateRemoteAgentRequest) (*agentsv1.UpdateRemoteAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateRemoteAgentURL(req.GetRemoteAgent()); err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetRemoteAgent(ctx, wsID, req.GetRemoteAgent().GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("updating remote agent", "workspace_id", wsID, "id", req.GetRemoteAgent().GetId(), "name", req.GetRemoteAgent().GetName())

	r, err := mutateWithRuntime(
		func() (*agentsv1.RemoteAgent, error) {
			return s.repo.UpdateRemoteAgent(ctx, wsID, req.GetRemoteAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateRemoteAgent(ctx, wsID, proto.Clone(prev).(*agentsv1.RemoteAgent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("update remote agent failed", "workspace_id", wsID, "id", req.GetRemoteAgent().GetId(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("remote agent updated", "workspace_id", wsID, "id", r.GetId(), "name", r.GetName())
	return &agentsv1.UpdateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) DeleteRemoteAgent(ctx context.Context, req *agentsv1.DeleteRemoteAgentRequest) (*agentsv1.DeleteRemoteAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetRemoteAgent(ctx, wsID, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("deleting remote agent", "workspace_id", wsID, "id", req.GetId(), "name", prev.GetName())

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteRemoteAgent(ctx, wsID, req.GetId())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateRemoteAgent(ctx, wsID, proto.Clone(prev).(*agentsv1.RemoteAgent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("delete remote agent failed", "workspace_id", wsID, "id", req.GetId(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("remote agent deleted", "workspace_id", wsID, "id", req.GetId(), "name", prev.GetName())
	return &agentsv1.DeleteRemoteAgentResponse{}, nil
}

func (s *RemoteAgentServiceServer) GetRemoteAgentStatus(ctx context.Context, req *agentsv1.GetRemoteAgentStatusRequest) (*agentsv1.GetRemoteAgentStatusResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	ra, err := s.repo.GetRemoteAgent(ctx, wsID, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}

	status := &agentsv1.RemoteAgentStatus{
		Id:        ra.GetId(),
		Protocol:  ra.GetProtocol(),
		CheckedAt: timestamppb.New(time.Now().UTC()),
	}

	switch ra.GetProtocol() {
	case agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON:
		cap := ra.GetDaemonCapability()
		if cap == "" {
			status.State = agentsv1.RemoteAgentStatus_STATE_ERROR
			status.Detail = "daemon_capability is required for DAEMON protocol"
			return &agentsv1.GetRemoteAgentStatusResponse{Status: status}, nil
		}
		if s.daemonReg == nil {
			status.State = agentsv1.RemoteAgentStatus_STATE_UNSPECIFIED
			status.Detail = "daemon registry not wired"
			return &agentsv1.GetRemoteAgentStatusResponse{Status: status}, nil
		}
		conn := s.daemonReg.FindByCapability(cap)
		if conn == nil {
			status.State = agentsv1.RemoteAgentStatus_STATE_UNREACHABLE
			status.Detail = fmt.Sprintf("no online daemon serves capability %q", cap)
			return &agentsv1.GetRemoteAgentStatusResponse{Status: status}, nil
		}
		status.ServingDaemonId = conn.Info.GetDaemonId()
		if conn.ActiveTaskCount() > 0 {
			status.State = agentsv1.RemoteAgentStatus_STATE_ACTIVE
		} else {
			status.State = agentsv1.RemoteAgentStatus_STATE_IDLE
		}
		return &agentsv1.GetRemoteAgentStatusResponse{Status: status}, nil

	case agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A:
		if strings.TrimSpace(ra.GetUrl()) == "" {
			status.State = agentsv1.RemoteAgentStatus_STATE_ERROR
			status.Detail = "url is required for A2A protocol"
			return &agentsv1.GetRemoteAgentStatusResponse{Status: status}, nil
		}
		probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		start := time.Now()
		state, detail := probeA2AAgent(probeCtx, ra.GetUrl())
		status.LatencyMs = time.Since(start).Milliseconds()
		status.State = state
		status.Detail = detail
		return &agentsv1.GetRemoteAgentStatusResponse{Status: status}, nil

	default:
		status.State = agentsv1.RemoteAgentStatus_STATE_CONFIGURED
		status.Detail = fmt.Sprintf("protocol %s not probed", ra.GetProtocol().String())
		return &agentsv1.GetRemoteAgentStatusResponse{Status: status}, nil
	}
}

// validateRemoteAgentURL enforces an absolute http(s) URL on the A2A
// endpoint to prevent SSRF via arbitrary schemes or empty hosts. DAEMON
// protocol agents ignore the URL field, but if a value is supplied we still
// require it to be a valid URL so misconfigurations fail loudly.
func validateRemoteAgentURL(ra *agentsv1.RemoteAgent) error {
	raw := strings.TrimSpace(ra.GetUrl())
	if ra.GetProtocol() == agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A {
		if raw == "" {
			return twirp.RequiredArgumentError("url")
		}
		return validateHTTPURL("url", raw)
	}
	if raw == "" {
		return nil
	}
	return validateHTTPURL("url", raw)
}

func probeA2AAgent(ctx context.Context, baseURL string) (agentsv1.RemoteAgentStatus_State, string) {
	url := strings.TrimRight(baseURL, "/") + "/.well-known/agent.json"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return agentsv1.RemoteAgentStatus_STATE_ERROR, err.Error()
	}
	httpReq.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return agentsv1.RemoteAgentStatus_STATE_UNREACHABLE, err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return agentsv1.RemoteAgentStatus_STATE_ACTIVE, ""
	}
	return agentsv1.RemoteAgentStatus_STATE_UNREACHABLE, fmt.Sprintf("agent card returned %d", resp.StatusCode)
}

func (s *RemoteAgentServiceServer) reloadRuntime(ctx context.Context) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.ReloadRunner(ctx); err != nil {
		return toTwirpError(err)
	}
	return nil
}
