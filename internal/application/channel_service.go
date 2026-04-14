package application

import (
	"context"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type ChannelServiceServer struct {
	repo configrepo.ChannelRepository
}

func NewChannelServiceServer(repo configrepo.ChannelRepository) *ChannelServiceServer {
	return &ChannelServiceServer{repo: repo}
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
	c, err := s.repo.CreateChannel(ctx, req.GetChannel())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) UpdateChannel(ctx context.Context, req *agentsv1.UpdateChannelRequest) (*agentsv1.UpdateChannelResponse, error) {
	c, err := s.repo.UpdateChannel(ctx, req.GetChannel())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateChannelResponse{Channel: c}, nil
}

func (s *ChannelServiceServer) DeleteChannel(ctx context.Context, req *agentsv1.DeleteChannelRequest) (*agentsv1.DeleteChannelResponse, error) {
	if err := s.repo.DeleteChannel(ctx, req.GetName()); err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteChannelResponse{}, nil
}
