package application

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/workspace"
)

func requireWorkspace(ctx context.Context) (string, error) {
	id, ok := workspace.FromContext(ctx)
	if !ok {
		return "", connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace required (set X-Workspace-ID header)"))
	}
	return id, nil
}

func withWorkspace(ctx context.Context, id string) context.Context {
	return workspace.WithID(ctx, id)
}
