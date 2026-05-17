package application

import (
	"context"
	"fmt"
	"time"

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

// SetChannelManager wires the runtime channel manager so status queries can
// observe live poller state.
func (s *ChannelServiceServer) SetChannelManager(m ChannelStatusProvider) {
	s.manager = m
}

func (s *ChannelServiceServer) ListChannels(ctx context.Context, _ *agentsv1.ListChannelsRequest) (*agentsv1.ListChannelsResponse, error) {
	channels, err := s.repo.ListChannels(ctx)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListChannelsResponse{Channels: channels}, nil
}

func (s *ChannelServiceServer) GetChannel(ctx context.Context, req *agentsv1.GetChannelRequest) (*agentsv1.GetChannelResponse, error) {
	c, err := s.repo.GetChannel(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) CreateChannel(ctx context.Context, req *agentsv1.CreateChannelRequest) (*agentsv1.CreateChannelResponse, error) {
	c, err := mutateWithRuntime(
		func() (*agentsv1.AgentChannel, error) {
			return s.repo.CreateChannel(ctx, req.GetChannel())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteChannel(ctx, req.GetChannel().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) UpdateChannel(ctx context.Context, req *agentsv1.UpdateChannelRequest) (*agentsv1.UpdateChannelResponse, error) {
	prev, err := s.repo.GetChannel(ctx, req.GetChannel().GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	c, err := mutateWithRuntime(
		func() (*agentsv1.AgentChannel, error) {
			return s.repo.UpdateChannel(ctx, req.GetChannel())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateChannel(ctx, proto.Clone(prev).(*agentsv1.AgentChannel)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) DeleteChannel(ctx context.Context, req *agentsv1.DeleteChannelRequest) (*agentsv1.DeleteChannelResponse, error) {
	prev, err := s.repo.GetChannel(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteChannel(ctx, req.GetName())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateChannel(ctx, proto.Clone(prev).(*agentsv1.AgentChannel)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteChannelResponse{}, nil
}

func (s *ChannelServiceServer) GetChannelStatus(ctx context.Context, req *agentsv1.GetChannelStatusRequest) (*agentsv1.GetChannelStatusResponse, error) {
	c, err := s.repo.GetChannel(ctx, req.GetName())
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
	resp := &agentsv1.RestartChannelResponse{}
	if name := req.GetName(); name != "" {
		c, err := s.repo.GetChannel(ctx, name)
		if err != nil {
			return nil, toTwirpError(err)
		}
		resp.Channel = c
	}
	if err := s.reloadRuntime(ctx); err != nil {
		return nil, toTwirpError(err)
	}
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
	prev, err := s.repo.GetChannel(ctx, name)
	if err != nil {
		return nil, toTwirpError(err)
	}
	if prev.GetEnabled() == enabled {
		return prev, nil
	}

	next := proto.Clone(prev).(*agentsv1.AgentChannel)
	next.Enabled = enabled

	updated, err := mutateWithRuntime(
		func() (*agentsv1.AgentChannel, error) {
			return s.repo.UpdateChannel(ctx, next)
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateChannel(ctx, proto.Clone(prev).(*agentsv1.AgentChannel)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
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
