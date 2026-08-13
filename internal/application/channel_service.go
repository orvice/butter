package application

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ChannelServiceServer serves the legacy generic AgentChannel API after the
// Telegram cutover (issue #264/#273).
//
// Reads remain so an operator can still see what is in the database and
// migrate deliberately. Every mutation and every runtime action is refused:
// nothing starts these records any more, so accepting a write would store
// configuration that looks live and does nothing — which is precisely the
// failure the cutover exists to end.
//
// Telegram lives on TelegramChannelService / TelegramDestinationService.
// Discord is unsupported in this release pending a redesign.
type ChannelServiceServer struct {
	repo      configrepo.ChannelRepository
	agentRepo configrepo.AgentRepository
}

func NewChannelServiceServer(repo configrepo.ChannelRepository) *ChannelServiceServer {
	return &ChannelServiceServer{repo: repo}
}

// SetAgentRepo wires agent lookups used when rendering legacy records.
func (s *ChannelServiceServer) SetAgentRepo(repo configrepo.AgentRepository) {
	s.agentRepo = repo
}

// unsupported is the single refusal every legacy mutation returns.
//
// Unimplemented rather than FailedPrecondition: this is not a state the
// caller can fix by changing the request, it is a capability that no longer
// exists at this endpoint.
func unsupportedLegacyChannel(action string) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New(
		action+" is no longer supported: create a Telegram Channel and Destination instead "+
			"(agents.v1.TelegramChannelService / agents.v1.TelegramDestinationService). "+
			"Discord is unsupported in this release."))
}

func (s *ChannelServiceServer) ListChannels(ctx context.Context, _ *connect.Request[agentsv1.ListChannelsRequest]) (*connect.Response[agentsv1.ListChannelsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	channels, err := s.repo.ListChannels(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.ListChannelsResponse{Channels: channels}), nil
}

func (s *ChannelServiceServer) GetChannel(ctx context.Context, req *connect.Request[agentsv1.GetChannelRequest]) (*connect.Response[agentsv1.GetChannelResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	ch, err := s.repo.GetChannel(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.GetChannelResponse{Channel: ch}), nil
}

// GetChannelStatus reports that the record exists but does not run.
func (s *ChannelServiceServer) GetChannelStatus(ctx context.Context, req *connect.Request[agentsv1.GetChannelStatusRequest]) (*connect.Response[agentsv1.GetChannelStatusResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	ch, err := s.repo.GetChannel(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.GetChannelStatusResponse{
		Status: &agentsv1.ChannelStatus{
			Name:  ch.GetName(),
			State: agentsv1.ChannelStatus_STATE_ERROR,
			Detail: "legacy agent channels are not started in this release; " +
				"recreate this as a Telegram Channel and Destination",
		},
	}), nil
}

func (s *ChannelServiceServer) CreateChannel(context.Context, *connect.Request[agentsv1.CreateChannelRequest]) (*connect.Response[agentsv1.CreateChannelResponse], error) {
	return nil, unsupportedLegacyChannel("creating a generic agent channel")
}

func (s *ChannelServiceServer) UpdateChannel(context.Context, *connect.Request[agentsv1.UpdateChannelRequest]) (*connect.Response[agentsv1.UpdateChannelResponse], error) {
	return nil, unsupportedLegacyChannel("updating a generic agent channel")
}

// DeleteChannel remains available: removing a record that no longer runs is
// exactly the cleanup an operator needs after migrating.
func (s *ChannelServiceServer) DeleteChannel(ctx context.Context, req *connect.Request[agentsv1.DeleteChannelRequest]) (*connect.Response[agentsv1.DeleteChannelResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.repo.DeleteChannel(ctx, wsID, req.Msg.GetName()); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.DeleteChannelResponse{}), nil
}

func (s *ChannelServiceServer) RestartChannel(context.Context, *connect.Request[agentsv1.RestartChannelRequest]) (*connect.Response[agentsv1.RestartChannelResponse], error) {
	return nil, unsupportedLegacyChannel("restarting a generic agent channel")
}

func (s *ChannelServiceServer) PauseChannel(context.Context, *connect.Request[agentsv1.PauseChannelRequest]) (*connect.Response[agentsv1.PauseChannelResponse], error) {
	return nil, unsupportedLegacyChannel("pausing a generic agent channel")
}

func (s *ChannelServiceServer) ResumeChannel(context.Context, *connect.Request[agentsv1.ResumeChannelRequest]) (*connect.Response[agentsv1.ResumeChannelResponse], error) {
	return nil, unsupportedLegacyChannel("resuming a generic agent channel")
}
