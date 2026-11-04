package application

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"github.com/google/uuid"
	"go.orx.me/apps/butter/internal/repo/auth"
	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"slices"
	"strings"
)

type WorkStationHostServiceServer struct{ repo ws.Repository }

func NewWorkStationHostServiceServer() *WorkStationHostServiceServer {
	return &WorkStationHostServiceServer{}
}
func (s *WorkStationHostServiceServer) SetRepo(repo ws.Repository) { s.repo = repo }
func (s *WorkStationHostServiceServer) authorize(ctx context.Context, manage bool) error {
	if _, ok := auth.UserFromContext(ctx); !ok && !auth.IsAdmin(ctx) {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("sign in to use Work Station"))
	}
	if manage {
		if err := requireGlobalAdmin(ctx); err != nil {
			return err
		}
	}
	if s.repo == nil {
		return stationError(ws.ErrUnsupported)
	}
	return nil
}
func (s *WorkStationHostServiceServer) ListHosts(ctx context.Context, _ *connect.Request[v1.WorkStationHostServiceListHostsRequest]) (*connect.Response[v1.WorkStationHostServiceListHostsResponse], error) {
	if err := s.authorize(ctx, false); err != nil {
		return nil, err
	}
	hosts, err := s.repo.ListHosts(ctx)
	if err != nil {
		return nil, stationError(err)
	}
	slices.SortFunc(hosts, func(a, b ws.Host) int { return strings.Compare(a.ID, b.ID) })
	out := &v1.WorkStationHostServiceListHostsResponse{}
	for _, h := range hosts {
		out.Hosts = append(out.Hosts, h.Public())
	}
	return connect.NewResponse(out), nil
}
func (s *WorkStationHostServiceServer) GetHost(ctx context.Context, req *connect.Request[v1.WorkStationHostServiceGetHostRequest]) (*connect.Response[v1.WorkStationHostServiceGetHostResponse], error) {
	if err := s.authorize(ctx, false); err != nil {
		return nil, err
	}
	h, err := s.repo.GetHost(ctx, req.Msg.Id)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationHostServiceGetHostResponse{Host: h.Public()}), nil
}
func hostFromProto(h *v1.WorkStationHost) ws.Host {
	return ws.Host{ID: h.GetId(), Name: h.GetName(), APIBaseURL: h.GetApiBaseUrl(), WebBaseURL: h.GetWebBaseUrl(), Profile: h.GetProfile(), Revision: h.GetRevision()}
}
func (s *WorkStationHostServiceServer) CreateHost(ctx context.Context, req *connect.Request[v1.WorkStationHostServiceCreateHostRequest]) (*connect.Response[v1.WorkStationHostServiceCreateHostResponse], error) {
	if err := s.authorize(ctx, true); err != nil {
		return nil, err
	}
	h := hostFromProto(req.Msg.Host)
	h.ID = uuid.NewString()
	h.Revision = 1
	if err := ws.ValidateHost(&h); err != nil {
		return nil, stationError(err)
	}
	if err := s.repo.SaveHost(ctx, h, 0); err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationHostServiceCreateHostResponse{Host: h.Public()}), nil
}
func (s *WorkStationHostServiceServer) UpdateHost(ctx context.Context, req *connect.Request[v1.WorkStationHostServiceUpdateHostRequest]) (*connect.Response[v1.WorkStationHostServiceUpdateHostResponse], error) {
	if err := s.authorize(ctx, true); err != nil {
		return nil, err
	}
	h := hostFromProto(req.Msg.Host)
	if h.ID == ws.PlaneCloudHost || req.Msg.ExpectedRevision < 1 {
		return nil, stationError(ws.ErrInvalid)
	}
	old, err := s.repo.GetHost(ctx, h.ID)
	if err != nil {
		return nil, stationError(err)
	}
	if old.Revision != req.Msg.ExpectedRevision {
		return nil, stationError(ws.ErrConflict)
	}
	if err := ws.ValidateHost(&h); err != nil {
		return nil, stationError(err)
	}
	if h.APIBaseURL != old.APIBaseURL || h.WebBaseURL != old.WebBaseURL || h.Profile != old.Profile {
		used, err := s.repo.HostInUse(ctx, h.ID)
		if err != nil {
			return nil, stationError(err)
		}
		if used {
			return nil, stationError(ws.ErrConflict)
		}
	}
	h.Revision = req.Msg.ExpectedRevision + 1
	if err := s.repo.SaveHost(ctx, h, req.Msg.ExpectedRevision); err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationHostServiceUpdateHostResponse{Host: h.Public()}), nil
}
func (s *WorkStationHostServiceServer) DeleteHost(ctx context.Context, req *connect.Request[v1.WorkStationHostServiceDeleteHostRequest]) (*connect.Response[v1.WorkStationHostServiceDeleteHostResponse], error) {
	if err := s.authorize(ctx, true); err != nil {
		return nil, err
	}
	if req.Msg.Id == "" || req.Msg.ExpectedRevision < 1 {
		return nil, stationError(ws.ErrInvalid)
	}
	if err := s.repo.DeleteHost(ctx, req.Msg.Id, req.Msg.ExpectedRevision); err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationHostServiceDeleteHostResponse{}), nil
}
