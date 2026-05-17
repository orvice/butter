package application

import (
	"context"
	"fmt"

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
	providers, err := s.repo.ListModelProviders(ctx)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListModelProvidersResponse{ModelProviders: providers}, nil
}

func (s *ModelProviderServiceServer) GetModelProvider(ctx context.Context, req *agentsv1.GetModelProviderRequest) (*agentsv1.GetModelProviderResponse, error) {
	provider, err := s.repo.GetModelProvider(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetModelProviderResponse{ModelProvider: provider}, nil
}

func (s *ModelProviderServiceServer) CreateModelProvider(ctx context.Context, req *agentsv1.CreateModelProviderRequest) (*agentsv1.CreateModelProviderResponse, error) {
	provider, err := mutateWithRuntime(
		func() (*agentsv1.ModelProvider, error) {
			return s.repo.CreateModelProvider(ctx, req.GetModelProvider())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteModelProvider(ctx, req.GetModelProvider().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateModelProviderResponse{ModelProvider: provider}, nil
}

func (s *ModelProviderServiceServer) UpdateModelProvider(ctx context.Context, req *agentsv1.UpdateModelProviderRequest) (*agentsv1.UpdateModelProviderResponse, error) {
	prev, err := s.repo.GetModelProvider(ctx, req.GetModelProvider().GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	provider, err := mutateWithRuntime(
		func() (*agentsv1.ModelProvider, error) {
			return s.repo.UpdateModelProvider(ctx, req.GetModelProvider())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateModelProvider(ctx, proto.Clone(prev).(*agentsv1.ModelProvider)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateModelProviderResponse{ModelProvider: provider}, nil
}

func (s *ModelProviderServiceServer) DeleteModelProvider(ctx context.Context, req *agentsv1.DeleteModelProviderRequest) (*agentsv1.DeleteModelProviderResponse, error) {
	prev, err := s.repo.GetModelProvider(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteModelProvider(ctx, req.GetName())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateModelProvider(ctx, proto.Clone(prev).(*agentsv1.ModelProvider)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
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
