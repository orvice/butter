package application

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/agentfile"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type AgentFileServiceServer struct {
	repo         agentfile.Repository
	maxFileBytes int64
}

func NewAgentFileServiceServer(repo agentfile.Repository) *AgentFileServiceServer {
	return &AgentFileServiceServer{repo: repo, maxFileBytes: defaultAgentFileMaxBytes}
}

func (s *AgentFileServiceServer) SetRepo(repo agentfile.Repository) {
	s.repo = repo
}

func (s *AgentFileServiceServer) SetMaxFileBytes(max int64) {
	if max <= 0 {
		max = defaultAgentFileMaxBytes
	}
	s.maxFileBytes = max
}

func (s *AgentFileServiceServer) requireRepo() (agentfile.Repository, error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("agent file repository not available"))
	}
	return s.repo, nil
}

func (s *AgentFileServiceServer) ListAgentFileSpaces(ctx context.Context, _ *connect.Request[agentsv1.ListAgentFileSpacesRequest]) (*connect.Response[agentsv1.ListAgentFileSpacesResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	spaces, err := repo.ListSpaces(ctx, wsID)
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.ListAgentFileSpacesResponse{Spaces: spaces}), nil
}

func (s *AgentFileServiceServer) GetAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.GetAgentFileSpaceRequest]) (*connect.Response[agentsv1.GetAgentFileSpaceResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	space, err := repo.GetSpace(ctx, wsID, req.Msg.GetId())
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.GetAgentFileSpaceResponse{Space: space}), nil
}

func (s *AgentFileServiceServer) CreateAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.CreateAgentFileSpaceRequest]) (*connect.Response[agentsv1.CreateAgentFileSpaceResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetSpace() == nil {
		return nil, connectx.RequiredArgument("space")
	}
	if req.Msg.GetSpace().GetName() == "" {
		return nil, connectx.RequiredArgument("space.name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating agent file space", "workspace_id", wsID, "name", req.Msg.GetSpace().GetName())
	space, err := repo.CreateSpace(ctx, wsID, req.Msg.GetSpace())
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.CreateAgentFileSpaceResponse{Space: space}), nil
}

func (s *AgentFileServiceServer) UpdateAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.UpdateAgentFileSpaceRequest]) (*connect.Response[agentsv1.UpdateAgentFileSpaceResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetSpace() == nil {
		return nil, connectx.RequiredArgument("space")
	}
	if req.Msg.GetSpace().GetId() == "" {
		return nil, connectx.RequiredArgument("space.id")
	}
	if req.Msg.GetSpace().GetName() == "" {
		return nil, connectx.RequiredArgument("space.name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	space, err := repo.UpdateSpace(ctx, wsID, req.Msg.GetSpace())
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.UpdateAgentFileSpaceResponse{Space: space}), nil
}

func (s *AgentFileServiceServer) DeleteAgentFileSpace(ctx context.Context, req *connect.Request[agentsv1.DeleteAgentFileSpaceRequest]) (*connect.Response[agentsv1.DeleteAgentFileSpaceResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeleteSpace(ctx, wsID, req.Msg.GetId()); err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.DeleteAgentFileSpaceResponse{}), nil
}

func (s *AgentFileServiceServer) ListAgentFiles(ctx context.Context, req *connect.Request[agentsv1.ListAgentFilesRequest]) (*connect.Response[agentsv1.ListAgentFilesResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetSpaceId() == "" {
		return nil, connectx.RequiredArgument("space_id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	files, err := repo.ListFiles(ctx, wsID, req.Msg.GetSpaceId(), req.Msg.GetPathPrefix())
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	sort.SliceStable(files, func(i, j int) bool { return files[i].GetPath() < files[j].GetPath() })
	return connect.NewResponse(&agentsv1.ListAgentFilesResponse{Files: files}), nil
}

func (s *AgentFileServiceServer) GetAgentFile(ctx context.Context, req *connect.Request[agentsv1.GetAgentFileRequest]) (*connect.Response[agentsv1.GetAgentFileResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetSpaceId() == "" {
		return nil, connectx.RequiredArgument("space_id")
	}
	if req.Msg.GetPath() == "" {
		return nil, connectx.RequiredArgument("path")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	file, content, err := repo.ReadFile(ctx, wsID, req.Msg.GetSpaceId(), req.Msg.GetPath(), req.Msg.GetVersion())
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.GetAgentFileResponse{File: file, Content: content}), nil
}

func (s *AgentFileServiceServer) WriteAgentFile(ctx context.Context, req *connect.Request[agentsv1.WriteAgentFileRequest]) (*connect.Response[agentsv1.WriteAgentFileResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetSpaceId() == "" {
		return nil, connectx.RequiredArgument("space_id")
	}
	if req.Msg.GetPath() == "" {
		return nil, connectx.RequiredArgument("path")
	}
	if max := s.maxFileBytes; max > 0 && int64(len([]byte(req.Msg.GetContent()))) > max {
		return nil, connectx.InvalidArgument("content", fmt.Sprintf("content exceeds max file size of %d bytes", max))
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	file, err := repo.WriteFile(ctx, wsID, req.Msg.GetSpaceId(), req.Msg.GetPath(), req.Msg.GetContent(), req.Msg.GetContentType(), req.Msg.GetMetadata())
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.WriteAgentFileResponse{File: file}), nil
}

func (s *AgentFileServiceServer) DeleteAgentFile(ctx context.Context, req *connect.Request[agentsv1.DeleteAgentFileRequest]) (*connect.Response[agentsv1.DeleteAgentFileResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetSpaceId() == "" {
		return nil, connectx.RequiredArgument("space_id")
	}
	if req.Msg.GetPath() == "" {
		return nil, connectx.RequiredArgument("path")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeleteFile(ctx, wsID, req.Msg.GetSpaceId(), req.Msg.GetPath()); err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.DeleteAgentFileResponse{}), nil
}

func (s *AgentFileServiceServer) SearchAgentFiles(ctx context.Context, req *connect.Request[agentsv1.SearchAgentFilesRequest]) (*connect.Response[agentsv1.SearchAgentFilesResponse], error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.Msg.GetSpaceId() == "" {
		return nil, connectx.RequiredArgument("space_id")
	}
	if req.Msg.GetQuery() == "" {
		return nil, connectx.RequiredArgument("query")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	results, err := repo.SearchFiles(ctx, wsID, req.Msg.GetSpaceId(), req.Msg.GetQuery(), int(req.Msg.GetLimit()))
	if err != nil {
		return nil, toConnectAgentFileError(err)
	}
	return connect.NewResponse(&agentsv1.SearchAgentFilesResponse{Results: results}), nil
}

func toConnectAgentFileError(err error) *connect.Error {
	var cerr *connect.Error
	if errors.As(err, &cerr) {
		return cerr
	}
	if errors.Is(err, agentfile.ErrNotFound) {
		return connectx.NotFound(err.Error())
	}
	if errors.Is(err, agentfile.ErrAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, errors.New(err.Error()))
	}
	if errors.Is(err, agentfile.ErrInvalidPath) {
		return connectx.InvalidArgument("path", err.Error())
	}
	return connectx.InternalWith(err)
}

const defaultAgentFileMaxBytes int64 = 256 * 1024
