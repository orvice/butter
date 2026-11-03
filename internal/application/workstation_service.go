package application

import (
	"context"
	"errors"
	"strings"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/auth"
	githost "go.orx.me/apps/butter/internal/repo/githost"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WorkStationServiceServer struct {
	service       *ws.Service
	workspaceRepo workspacerepo.Repository
}

func NewWorkStationServiceServer() *WorkStationServiceServer { return &WorkStationServiceServer{} }
func (s *WorkStationServiceServer) SetService(service *ws.Service, workspaces workspacerepo.Repository) {
	s.service = service
	s.workspaceRepo = workspaces
}

// WorkStationHostCheck only accepts the existing registry's official GitHub
// profile. No workspace input becomes an endpoint or a redirect target.
func WorkStationHostCheck(repo githost.Repository) func(context.Context, string) error {
	return func(ctx context.Context, id string) error {
		if repo == nil {
			return ws.ErrUnsupported
		}
		h, err := repo.Get(ctx, id)
		if err != nil {
			return ws.ErrUnsupported
		}
		if h.Kind != v1.GitHostKind_GIT_HOST_KIND_GITHUB || strings.TrimRight(h.ApiBaseUrl, "/") != "https://api.github.com" || strings.TrimRight(h.WebBaseUrl, "/") != "https://github.com" {
			return ws.ErrUnsupported
		}
		return nil
	}
}
func (s *WorkStationServiceServer) authorize(ctx context.Context, manage bool) (string, error) {
	w, err := requireWorkspace(ctx)
	if err != nil {
		return "", err
	}
	if auth.IsAdmin(ctx) {
		log.FromContext(ctx).Info("global admin accessing work station", "audit", "admin_work_station_access", "workspace_id", w)
	} else {
		u, ok := auth.UserFromContext(ctx)
		if !ok {
			return "", connect.NewError(connect.CodeUnauthenticated, errors.New("sign in to use Work Station"))
		}
		if s.workspaceRepo == nil {
			return "", connect.NewError(connect.CodeFailedPrecondition, errors.New("work station not configured"))
		}
		member, err := s.workspaceRepo.GetMember(ctx, w, u.Id)
		if errors.Is(err, workspacerepo.ErrNotFound) {
			return "", connect.NewError(connect.CodeNotFound, errors.New("workspace not found"))
		}
		if err != nil {
			return "", stationError(err)
		}
		if manage && member.Role != "owner" {
			return "", connect.NewError(connect.CodePermissionDenied, errors.New("workspace owner role required"))
		}
	}
	if s.service == nil {
		return "", connect.NewError(connect.CodeFailedPrecondition, errors.New("work station not configured"))
	}
	return w, nil
}
func stationError(err error) error {
	if err == nil {
		return nil
	}
	code := connect.CodeInternal
	message := "Work Station is temporarily unavailable"
	switch {
	case errors.Is(err, ws.ErrNotFound):
		code = connect.CodeNotFound
		message = ws.ErrNotFound.Error()
	case errors.Is(err, ws.ErrConflict), errors.Is(err, ws.ErrFenced):
		code = connect.CodeAborted
		message = ws.ErrConflict.Error()
	case errors.Is(err, ws.ErrInvalid):
		code = connect.CodeInvalidArgument
		message = err.Error()
	case errors.Is(err, ws.ErrUnsupported), errors.Is(err, ws.ErrPaused):
		code = connect.CodeFailedPrecondition
		message = err.Error()
	default:
		var f *ws.Failure
		if errors.As(err, &f) {
			code = connect.CodeUnavailable
			message = "Provider check failed: " + f.Category
			if f.Category == "AUTH_INVALID" || f.Category == "ACCESS_DENIED" {
				code = connect.CodeFailedPrecondition
			}
		}
	}
	return connect.NewError(code, errors.New(message))
}
func (s *WorkStationServiceServer) ListConnections(ctx context.Context, _ *connect.Request[v1.WorkStationServiceListConnectionsRequest]) (*connect.Response[v1.WorkStationServiceListConnectionsResponse], error) {
	w, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}
	connections, err := s.service.Connections(ctx, w)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceListConnectionsResponse{Connections: connections}), nil
}
func (s *WorkStationServiceServer) GetConnection(ctx context.Context, req *connect.Request[v1.WorkStationServiceGetConnectionRequest]) (*connect.Response[v1.WorkStationServiceGetConnectionResponse], error) {
	w, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}
	connection, err := s.service.Connection(ctx, w, req.Msg.Id)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceGetConnectionResponse{Connection: connection}), nil
}

func (s *WorkStationServiceServer) CreateConnection(ctx context.Context, req *connect.Request[v1.WorkStationServiceCreateConnectionRequest]) (*connect.Response[v1.WorkStationServiceCreateConnectionResponse], error) {
	w, err := s.authorize(ctx, true)
	if err != nil {
		return nil, err
	}
	r, err := s.service.Create(ctx, w, req.Msg.Name, req.Msg.HostId, req.Msg.ApiKey)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceCreateConnectionResponse{Connection: ws.Public(r)}), nil
}
func (s *WorkStationServiceServer) UpdateConnection(ctx context.Context, req *connect.Request[v1.WorkStationServiceUpdateConnectionRequest]) (*connect.Response[v1.WorkStationServiceUpdateConnectionResponse], error) {
	w, err := s.authorize(ctx, true)
	if err != nil {
		return nil, err
	}
	r, err := s.service.Update(ctx, w, req.Msg.Id, req.Msg.ExpectedRevision, req.Msg.Name, req.Msg.ApiKey)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceUpdateConnectionResponse{Connection: ws.Public(r)}), nil
}
func (s *WorkStationServiceServer) DeleteConnection(ctx context.Context, req *connect.Request[v1.WorkStationServiceDeleteConnectionRequest]) (*connect.Response[v1.WorkStationServiceDeleteConnectionResponse], error) {
	w, err := s.authorize(ctx, true)
	if err != nil {
		return nil, err
	}
	if err := s.service.Delete(ctx, w, req.Msg.Id, req.Msg.ExpectedRevision); err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceDeleteConnectionResponse{}), nil
}
func (s *WorkStationServiceServer) TestConnection(ctx context.Context, req *connect.Request[v1.WorkStationServiceTestConnectionRequest]) (*connect.Response[v1.WorkStationServiceTestConnectionResponse], error) {
	w, err := s.authorize(ctx, true)
	if err != nil {
		return nil, err
	}
	r, err := s.service.Test(ctx, w, req.Msg.Id)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceTestConnectionResponse{Connection: ws.Public(r)}), nil
}
func (s *WorkStationServiceServer) DiscoverScopes(ctx context.Context, req *connect.Request[v1.WorkStationServiceDiscoverScopesRequest]) (*connect.Response[v1.WorkStationServiceDiscoverScopesResponse], error) {
	w, err := s.authorize(ctx, true)
	if err != nil {
		return nil, err
	}
	page, err := s.service.Discover(ctx, w, req.Msg.Id, req.Msg.PageToken)
	if err != nil {
		return nil, stationError(err)
	}
	res := &v1.WorkStationServiceDiscoverScopesResponse{NextPageToken: page.Next}
	for _, sc := range page.Scopes {
		res.Scopes = append(res.Scopes, &v1.WorkStationScope{Id: sc.ID, Name: sc.Name})
	}
	return connect.NewResponse(res), nil
}
func (s *WorkStationServiceServer) SetSelectedScopes(ctx context.Context, req *connect.Request[v1.WorkStationServiceSetSelectedScopesRequest]) (*connect.Response[v1.WorkStationServiceSetSelectedScopesResponse], error) {
	w, err := s.authorize(ctx, true)
	if err != nil {
		return nil, err
	}
	r, err := s.service.Select(ctx, w, req.Msg.Id, req.Msg.ExpectedRevision, req.Msg.ScopeIds)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceSetSelectedScopesResponse{Connection: ws.Public(r)}), nil
}
func (s *WorkStationServiceServer) Refresh(ctx context.Context, req *connect.Request[v1.WorkStationServiceRefreshRequest]) (*connect.Response[v1.WorkStationServiceRefreshResponse], error) {
	w, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}
	status, due, err := s.service.Refresh(ctx, w, req.Msg.Id)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceRefreshResponse{Status: status, NextAttemptAt: timestamppb.New(due)}), nil
}
func (s *WorkStationServiceServer) ListItems(ctx context.Context, req *connect.Request[v1.WorkStationServiceListItemsRequest]) (*connect.Response[v1.WorkStationServiceListItemsResponse], error) {
	w, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}
	q := ws.Query{Kind: req.Msg.Kind, ConnectionID: req.Msg.ConnectionId, ScopeID: req.Msg.ScopeId, Status: req.Msg.Status, Assignee: req.Msg.Assignee, Keyword: req.Msg.Keyword, Limit: int(req.Msg.PageSize)}
	items, next, err := s.service.List(ctx, w, q, req.Msg.PageToken)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceListItemsResponse{Items: items, NextPageToken: next}), nil
}
func (s *WorkStationServiceServer) ListFilterOptions(ctx context.Context, req *connect.Request[v1.WorkStationServiceListFilterOptionsRequest]) (*connect.Response[v1.WorkStationServiceListFilterOptionsResponse], error) {
	w, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}
	options, err := s.service.FilterOptions(ctx, w, ws.Query{Kind: req.Msg.Kind, ConnectionID: req.Msg.ConnectionId, ScopeID: req.Msg.ScopeId})
	if err != nil {
		return nil, stationError(err)
	}
	toProto := func(values []string) []*v1.WorkStationFilterOption {
		out := make([]*v1.WorkStationFilterOption, 0, len(values))
		for _, value := range values {
			out = append(out, &v1.WorkStationFilterOption{Value: value, Label: value})
		}
		return out
	}
	return connect.NewResponse(&v1.WorkStationServiceListFilterOptionsResponse{
		Statuses:  toProto(options.Statuses),
		Assignees: toProto(options.Assignees),
		Labels:    toProto(options.Labels),
	}), nil
}

func (s *WorkStationServiceServer) GetItem(ctx context.Context, req *connect.Request[v1.WorkStationServiceGetItemRequest]) (*connect.Response[v1.WorkStationServiceGetItemResponse], error) {
	w, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}
	item, err := s.service.Get(ctx, w, req.Msg.Id)
	if err != nil {
		return nil, stationError(err)
	}
	return connect.NewResponse(&v1.WorkStationServiceGetItemResponse{Item: item}), nil
}
