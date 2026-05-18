package application

import (
	"context"
	"fmt"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/channel"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChannelStatusProvider abstracts channel.Manager for unit-testing.
type ChannelStatusProvider interface {
	ChannelStatus(ctx context.Context, name string) (channel.RuntimeState, string, error)
}

type ChannelServiceServer struct {
	repo    configrepo.ChannelRepository
	runtime ConfigRuntime
	manager ChannelStatusProvider
}

func NewChannelServiceServer(repo configrepo.ChannelRepository) *ChannelServiceServer {
	return &ChannelServiceServer{repo: repo}
}

func (s *ChannelServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

func (s *ChannelServiceServer) SetChannelManager(m ChannelStatusProvider) {
	s.manager = m
}

func (s *ChannelServiceServer) ListChannels(ctx context.Context, _ *agentsv1.ListChannelsRequest) (*agentsv1.ListChannelsResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	channels, err := s.repo.ListChannels(ctx, wsID)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListChannelsResponse{Channels: channels}, nil
}

func (s *ChannelServiceServer) GetChannel(ctx context.Context, req *agentsv1.GetChannelRequest) (*agentsv1.GetChannelResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.GetChannel(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) CreateChannel(ctx context.Context, req *agentsv1.CreateChannelRequest) (*agentsv1.CreateChannelResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating channel",
		"workspace_id", wsID,
		"channel", req.GetChannel().GetName(),
		"platform", req.GetChannel().GetPlatform(),
		"enabled", req.GetChannel().GetEnabled(),
	)
	c, err := mutateWithRuntime(
		func() (*agentsv1.AgentChannel, error) {
			return s.repo.CreateChannel(ctx, wsID, req.GetChannel())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteChannel(ctx, wsID, req.GetChannel().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("create channel failed", "workspace_id", wsID, "channel", req.GetChannel().GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("channel created", "workspace_id", wsID, "channel", c.GetName(), "platform", c.GetPlatform())
	return &agentsv1.CreateChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) UpdateChannel(ctx context.Context, req *agentsv1.UpdateChannelRequest) (*agentsv1.UpdateChannelResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetChannel(ctx, wsID, req.GetChannel().GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("updating channel", "workspace_id", wsID, "channel", req.GetChannel().GetName())

	c, err := mutateWithRuntime(
		func() (*agentsv1.AgentChannel, error) {
			return s.repo.UpdateChannel(ctx, wsID, req.GetChannel())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateChannel(ctx, wsID, proto.Clone(prev).(*agentsv1.AgentChannel)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("update channel failed", "workspace_id", wsID, "channel", req.GetChannel().GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("channel updated", "workspace_id", wsID, "channel", c.GetName())
	return &agentsv1.UpdateChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) DeleteChannel(ctx context.Context, req *agentsv1.DeleteChannelRequest) (*agentsv1.DeleteChannelResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetChannel(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("deleting channel", "workspace_id", wsID, "channel", req.GetName())

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteChannel(ctx, wsID, req.GetName())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateChannel(ctx, wsID, proto.Clone(prev).(*agentsv1.AgentChannel)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("delete channel failed", "workspace_id", wsID, "channel", req.GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("channel deleted", "workspace_id", wsID, "channel", req.GetName())
	return &agentsv1.DeleteChannelResponse{}, nil
}

func (s *ChannelServiceServer) GetChannelStatus(ctx context.Context, req *agentsv1.GetChannelStatusRequest) (*agentsv1.GetChannelStatusResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.GetChannel(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	state := agentsv1.ChannelStatus_STATE_UNSPECIFIED
	detail := ""
	switch {
	case !c.GetEnabled():
		state = agentsv1.ChannelStatus_STATE_DISABLED
		detail = "channel disabled in config"
	case s.manager == nil:
		state = agentsv1.ChannelStatus_STATE_UNSPECIFIED
		detail = "channel manager not wired"
	default:
		rs, d, err := s.manager.ChannelStatus(ctx, c.GetName())
		if err != nil {
			return nil, twirp.InternalErrorWith(fmt.Errorf("channel status: %w", err))
		}
		detail = d
		switch rs {
		case channel.RuntimeStateLive:
			state = agentsv1.ChannelStatus_STATE_LIVE
		case channel.RuntimeStateDisabled:
			state = agentsv1.ChannelStatus_STATE_DISABLED
		case channel.RuntimeStateUnsupported:
			state = agentsv1.ChannelStatus_STATE_ERROR
		case channel.RuntimeStateNotFound:
			return nil, twirp.NotFoundError("channel not found")
		}
	}

	status := &agentsv1.ChannelStatus{
		Name:       c.GetName(),
		Platform:   c.GetPlatform(),
		State:      state,
		Detail:     detail,
		LastPollAt: timestamppb.New(time.Now().UTC()),
	}
	return &agentsv1.GetChannelStatusResponse{Status: status}, nil
}

func (s *ChannelServiceServer) RestartChannel(ctx context.Context, req *agentsv1.RestartChannelRequest) (*agentsv1.RestartChannelResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("restarting channel", "workspace_id", wsID, "channel", req.GetName())
	resp := &agentsv1.RestartChannelResponse{}
	if name := req.GetName(); name != "" {
		c, err := s.repo.GetChannel(ctx, wsID, name)
		if err != nil {
			return nil, toTwirpError(err)
		}
		resp.Channel = c
	}
	if err := s.reloadRuntime(ctx); err != nil {
		logger.Error("restart channel: reload runtime failed", "workspace_id", wsID, "channel", req.GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("channel restarted", "workspace_id", wsID, "channel", req.GetName())
	return resp, nil
}

func (s *ChannelServiceServer) PauseChannel(ctx context.Context, req *agentsv1.PauseChannelRequest) (*agentsv1.PauseChannelResponse, error) {
	c, err := s.toggleChannelEnabled(ctx, req.GetName(), false)
	if err != nil {
		return nil, err
	}
	return &agentsv1.PauseChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) ResumeChannel(ctx context.Context, req *agentsv1.ResumeChannelRequest) (*agentsv1.ResumeChannelResponse, error) {
	c, err := s.toggleChannelEnabled(ctx, req.GetName(), true)
	if err != nil {
		return nil, err
	}
	return &agentsv1.ResumeChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) toggleChannelEnabled(ctx context.Context, name string, enabled bool) (*agentsv1.AgentChannel, error) {
	if name == "" {
		return nil, twirp.RequiredArgumentError("name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetChannel(ctx, wsID, name)
	if err != nil {
		return nil, toTwirpError(err)
	}
	if prev.GetEnabled() == enabled {
		logger.Debug("channel already in target state", "workspace_id", wsID, "channel", name, "enabled", enabled)
		return prev, nil
	}
	logger.Info("toggling channel enabled", "workspace_id", wsID, "channel", name, "enabled", enabled)

	next := proto.Clone(prev).(*agentsv1.AgentChannel)
	next.Enabled = enabled

	updated, err := mutateWithRuntime(
		func() (*agentsv1.AgentChannel, error) {
			return s.repo.UpdateChannel(ctx, wsID, next)
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateChannel(ctx, wsID, proto.Clone(prev).(*agentsv1.AgentChannel)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("toggle channel enabled failed", "workspace_id", wsID, "channel", name, "enabled", enabled, "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("channel enabled toggled", "workspace_id", wsID, "channel", name, "enabled", enabled)
	return updated, nil
}

func (s *ChannelServiceServer) reloadRuntime(ctx context.Context) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.ReloadChannels(ctx); err != nil {
		return toTwirpError(fmt.Errorf("reload channels: %w", err))
	}
	return nil
}
