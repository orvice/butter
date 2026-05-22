package agentfile

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidPath   = errors.New("invalid path")
)

// Repository stores workspace-scoped agent file spaces and text files.
type Repository interface {
	EnsureIndexes(ctx context.Context) error

	ListSpaces(ctx context.Context, workspaceID string) ([]*agentsv1.AgentFileSpace, error)
	GetSpace(ctx context.Context, workspaceID, id string) (*agentsv1.AgentFileSpace, error)
	CreateSpace(ctx context.Context, workspaceID string, space *agentsv1.AgentFileSpace) (*agentsv1.AgentFileSpace, error)
	UpdateSpace(ctx context.Context, workspaceID string, space *agentsv1.AgentFileSpace) (*agentsv1.AgentFileSpace, error)
	DeleteSpace(ctx context.Context, workspaceID, id string) error

	ListFiles(ctx context.Context, workspaceID, spaceID, pathPrefix string) ([]*agentsv1.AgentFile, error)
	GetFile(ctx context.Context, workspaceID, spaceID, path string) (*agentsv1.AgentFile, error)
	ReadFile(ctx context.Context, workspaceID, spaceID, path string, version int64) (*agentsv1.AgentFile, string, error)
	WriteFile(ctx context.Context, workspaceID, spaceID, path, content, contentType string, metadata map[string]string) (*agentsv1.AgentFile, error)
	DeleteFile(ctx context.Context, workspaceID, spaceID, path string) error
	SearchFiles(ctx context.Context, workspaceID, spaceID, query string, limit int) ([]*agentsv1.AgentFileSearchResult, error)
}
