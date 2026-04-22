package application

import (
	"context"
	"fmt"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

type ChannelServiceServer struct {
	repo    configrepo.ChannelRepository
	runtime ConfigRuntime
}

func NewChannelServiceServer(repo configrepo.ChannelRepository) *ChannelServiceServer {
	return &ChannelServiceServer{repo: repo}
}

func (s *ChannelServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
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

func (s *ChannelServiceServer) reloadRuntime(ctx context.Context) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.ReloadChannels(ctx); err != nil {
		return toTwirpError(fmt.Errorf("reload channels: %w", err))
	}
	return nil
}
