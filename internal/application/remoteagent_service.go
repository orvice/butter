package application

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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

// SetDaemonRegistry wires the daemon registry used by GetRemoteAgentStatus
// for DAEMON-protocol remote agents.
func (s *RemoteAgentServiceServer) SetDaemonRegistry(reg *daemon.Registry) {
	s.daemonReg = reg
}

func (s *RemoteAgentServiceServer) ListRemoteAgents(ctx context.Context, _ *agentsv1.ListRemoteAgentsRequest) (*agentsv1.ListRemoteAgentsResponse, error) {
	agents, err := s.repo.ListRemoteAgents(ctx)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListRemoteAgentsResponse{RemoteAgents: agents}, nil
}

func (s *RemoteAgentServiceServer) GetRemoteAgent(ctx context.Context, req *agentsv1.GetRemoteAgentRequest) (*agentsv1.GetRemoteAgentResponse, error) {
	r, err := s.repo.GetRemoteAgent(ctx, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) CreateRemoteAgent(ctx context.Context, req *agentsv1.CreateRemoteAgentRequest) (*agentsv1.CreateRemoteAgentResponse, error) {
	r, err := mutateWithRuntime(
		func() (*agentsv1.RemoteAgent, error) {
			return s.repo.CreateRemoteAgent(ctx, req.GetRemoteAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteRemoteAgent(ctx, req.GetRemoteAgent().GetId()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) UpdateRemoteAgent(ctx context.Context, req *agentsv1.UpdateRemoteAgentRequest) (*agentsv1.UpdateRemoteAgentResponse, error) {
	prev, err := s.repo.GetRemoteAgent(ctx, req.GetRemoteAgent().GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}

	r, err := mutateWithRuntime(
		func() (*agentsv1.RemoteAgent, error) {
			return s.repo.UpdateRemoteAgent(ctx, req.GetRemoteAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateRemoteAgent(ctx, proto.Clone(prev).(*agentsv1.RemoteAgent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) DeleteRemoteAgent(ctx context.Context, req *agentsv1.DeleteRemoteAgentRequest) (*agentsv1.DeleteRemoteAgentResponse, error) {
	prev, err := s.repo.GetRemoteAgent(ctx, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteRemoteAgent(ctx, req.GetId())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateRemoteAgent(ctx, proto.Clone(prev).(*agentsv1.RemoteAgent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteRemoteAgentResponse{}, nil
}

func (s *RemoteAgentServiceServer) GetRemoteAgentStatus(ctx context.Context, req *agentsv1.GetRemoteAgentStatusRequest) (*agentsv1.GetRemoteAgentStatusResponse, error) {
	ra, err := s.repo.GetRemoteAgent(ctx, req.GetId())
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

// probeA2AAgent issues an HTTP GET to the A2A agent card endpoint and reports
// the resulting state.
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
