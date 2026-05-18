package application

import (
	"context"
	"fmt"

	"butterfly.orx.me/core/log"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

type ModelProviderServiceServer struct {
	repo    configrepo.ModelProviderRepository
	runtime ConfigRuntime
}

func NewModelProviderServiceServer(repo configrepo.ModelProviderRepository) *ModelProviderServiceServer {
	return &ModelProviderServiceServer{repo: repo}
}

func (s *ModelProviderServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

func (s *ModelProviderServiceServer) ListModelProviders(ctx context.Context, _ *agentsv1.ListModelProvidersRequest) (*agentsv1.ListModelProvidersResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	providers, err := s.repo.ListModelProviders(ctx, wsID)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListModelProvidersResponse{ModelProviders: providers}, nil
}

func (s *ModelProviderServiceServer) GetModelProvider(ctx context.Context, req *agentsv1.GetModelProviderRequest) (*agentsv1.GetModelProviderResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := s.repo.GetModelProvider(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetModelProviderResponse{ModelProvider: provider}, nil
}

func (s *ModelProviderServiceServer) CreateModelProvider(ctx context.Context, req *agentsv1.CreateModelProviderRequest) (*agentsv1.CreateModelProviderResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating model provider", "workspace_id", wsID, "name", req.GetModelProvider().GetName())
	provider, err := mutateWithRuntime(
		func() (*agentsv1.ModelProvider, error) {
			return s.repo.CreateModelProvider(ctx, wsID, req.GetModelProvider())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteModelProvider(ctx, wsID, req.GetModelProvider().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("create model provider failed", "workspace_id", wsID, "name", req.GetModelProvider().GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("model provider created", "workspace_id", wsID, "name", provider.GetName())
	return &agentsv1.CreateModelProviderResponse{ModelProvider: provider}, nil
}

func (s *ModelProviderServiceServer) UpdateModelProvider(ctx context.Context, req *agentsv1.UpdateModelProviderRequest) (*agentsv1.UpdateModelProviderResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetModelProvider(ctx, wsID, req.GetModelProvider().GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("updating model provider", "workspace_id", wsID, "name", req.GetModelProvider().GetName())

	provider, err := mutateWithRuntime(
		func() (*agentsv1.ModelProvider, error) {
			return s.repo.UpdateModelProvider(ctx, wsID, req.GetModelProvider())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateModelProvider(ctx, wsID, proto.Clone(prev).(*agentsv1.ModelProvider)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("update model provider failed", "workspace_id", wsID, "name", req.GetModelProvider().GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("model provider updated", "workspace_id", wsID, "name", provider.GetName())
	return &agentsv1.UpdateModelProviderResponse{ModelProvider: provider}, nil
}

func (s *ModelProviderServiceServer) DeleteModelProvider(ctx context.Context, req *agentsv1.DeleteModelProviderRequest) (*agentsv1.DeleteModelProviderResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetModelProvider(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("deleting model provider", "workspace_id", wsID, "name", req.GetName())

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteModelProvider(ctx, wsID, req.GetName())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateModelProvider(ctx, wsID, proto.Clone(prev).(*agentsv1.ModelProvider)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("delete model provider failed", "workspace_id", wsID, "name", req.GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("model provider deleted", "workspace_id", wsID, "name", req.GetName())
	return &agentsv1.DeleteModelProviderResponse{}, nil
}

func (s *ModelProviderServiceServer) reloadRuntime(ctx context.Context) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.ReloadRunner(ctx); err != nil {
		return toTwirpError(fmt.Errorf("reload model providers: %w", err))
	}
	return nil
}
