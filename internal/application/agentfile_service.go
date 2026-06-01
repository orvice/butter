package application

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"butterfly.orx.me/core/log"
	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/repo/agentfile"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type AgentFileServiceServer struct {
	agentsv1.UnimplementedAgentFileServiceServer

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
		return nil, twirp.NewError(twirp.FailedPrecondition, "agent file repository not available")
	}
	return s.repo, nil
}

func (s *AgentFileServiceServer) ListAgentFileSpaces(ctx context.Context, _ *agentsv1.ListAgentFileSpacesRequest) (*agentsv1.ListAgentFileSpacesResponse, error) {
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
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.ListAgentFileSpacesResponse{Spaces: spaces}, nil
}

func (s *AgentFileServiceServer) GetAgentFileSpace(ctx context.Context, req *agentsv1.GetAgentFileSpaceRequest) (*agentsv1.GetAgentFileSpaceResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, twirp.RequiredArgumentError("id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	space, err := repo.GetSpace(ctx, wsID, req.GetId())
	if err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.GetAgentFileSpaceResponse{Space: space}, nil
}

func (s *AgentFileServiceServer) CreateAgentFileSpace(ctx context.Context, req *agentsv1.CreateAgentFileSpaceRequest) (*agentsv1.CreateAgentFileSpaceResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetSpace() == nil {
		return nil, twirp.RequiredArgumentError("space")
	}
	if req.GetSpace().GetName() == "" {
		return nil, twirp.RequiredArgumentError("space.name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating agent file space", "workspace_id", wsID, "name", req.GetSpace().GetName())
	space, err := repo.CreateSpace(ctx, wsID, req.GetSpace())
	if err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.CreateAgentFileSpaceResponse{Space: space}, nil
}

func (s *AgentFileServiceServer) UpdateAgentFileSpace(ctx context.Context, req *agentsv1.UpdateAgentFileSpaceRequest) (*agentsv1.UpdateAgentFileSpaceResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetSpace() == nil {
		return nil, twirp.RequiredArgumentError("space")
	}
	if req.GetSpace().GetId() == "" {
		return nil, twirp.RequiredArgumentError("space.id")
	}
	if req.GetSpace().GetName() == "" {
		return nil, twirp.RequiredArgumentError("space.name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	space, err := repo.UpdateSpace(ctx, wsID, req.GetSpace())
	if err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.UpdateAgentFileSpaceResponse{Space: space}, nil
}

func (s *AgentFileServiceServer) DeleteAgentFileSpace(ctx context.Context, req *agentsv1.DeleteAgentFileSpaceRequest) (*agentsv1.DeleteAgentFileSpaceResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, twirp.RequiredArgumentError("id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeleteSpace(ctx, wsID, req.GetId()); err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.DeleteAgentFileSpaceResponse{}, nil
}

func (s *AgentFileServiceServer) ListAgentFiles(ctx context.Context, req *agentsv1.ListAgentFilesRequest) (*agentsv1.ListAgentFilesResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetSpaceId() == "" {
		return nil, twirp.RequiredArgumentError("space_id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	files, err := repo.ListFiles(ctx, wsID, req.GetSpaceId(), req.GetPathPrefix())
	if err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	sort.SliceStable(files, func(i, j int) bool { return files[i].GetPath() < files[j].GetPath() })
	return &agentsv1.ListAgentFilesResponse{Files: files}, nil
}

func (s *AgentFileServiceServer) GetAgentFile(ctx context.Context, req *agentsv1.GetAgentFileRequest) (*agentsv1.GetAgentFileResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetSpaceId() == "" {
		return nil, twirp.RequiredArgumentError("space_id")
	}
	if req.GetPath() == "" {
		return nil, twirp.RequiredArgumentError("path")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	file, content, err := repo.ReadFile(ctx, wsID, req.GetSpaceId(), req.GetPath(), req.GetVersion())
	if err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.GetAgentFileResponse{File: file, Content: content}, nil
}

func (s *AgentFileServiceServer) WriteAgentFile(ctx context.Context, req *agentsv1.WriteAgentFileRequest) (*agentsv1.WriteAgentFileResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetSpaceId() == "" {
		return nil, twirp.RequiredArgumentError("space_id")
	}
	if req.GetPath() == "" {
		return nil, twirp.RequiredArgumentError("path")
	}
	if max := s.maxFileBytes; max > 0 && int64(len([]byte(req.GetContent()))) > max {
		return nil, twirp.InvalidArgumentError("content", fmt.Sprintf("content exceeds max file size of %d bytes", max))
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	file, err := repo.WriteFile(ctx, wsID, req.GetSpaceId(), req.GetPath(), req.GetContent(), req.GetContentType(), req.GetMetadata())
	if err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.WriteAgentFileResponse{File: file}, nil
}

func (s *AgentFileServiceServer) DeleteAgentFile(ctx context.Context, req *agentsv1.DeleteAgentFileRequest) (*agentsv1.DeleteAgentFileResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetSpaceId() == "" {
		return nil, twirp.RequiredArgumentError("space_id")
	}
	if req.GetPath() == "" {
		return nil, twirp.RequiredArgumentError("path")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeleteFile(ctx, wsID, req.GetSpaceId(), req.GetPath()); err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.DeleteAgentFileResponse{}, nil
}

func (s *AgentFileServiceServer) SearchAgentFiles(ctx context.Context, req *agentsv1.SearchAgentFilesRequest) (*agentsv1.SearchAgentFilesResponse, error) {
	repo, err := s.requireRepo()
	if err != nil {
		return nil, err
	}
	if req.GetSpaceId() == "" {
		return nil, twirp.RequiredArgumentError("space_id")
	}
	if req.GetQuery() == "" {
		return nil, twirp.RequiredArgumentError("query")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	results, err := repo.SearchFiles(ctx, wsID, req.GetSpaceId(), req.GetQuery(), int(req.GetLimit()))
	if err != nil {
		return nil, toAgentFileTwirpError(err)
	}
	return &agentsv1.SearchAgentFilesResponse{Results: results}, nil
}

func toAgentFileTwirpError(err error) twirp.Error {
	if twerr, ok := err.(twirp.Error); ok {
		return twerr
	}
	if errors.Is(err, agentfile.ErrNotFound) {
		return twirp.NotFoundError(err.Error())
	}
	if errors.Is(err, agentfile.ErrAlreadyExists) {
		return twirp.NewError(twirp.AlreadyExists, err.Error())
	}
	if errors.Is(err, agentfile.ErrInvalidPath) {
		return twirp.InvalidArgumentError("path", err.Error())
	}
	return twirp.InternalErrorWith(err)
}

const defaultAgentFileMaxBytes int64 = 256 * 1024
