package planeinstallation

import (
	"context"
	"errors"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"time"
)

var (
	ErrNotFound = errors.New("plane installation not found")
	ErrConflict = errors.New("plane installation conflict")
)

type Tokens struct {
	AccessToken      string
	AccessTokenKeyID string
	ExpiresAt        time.Time
}
type Repository interface {
	EnsureIndexes(context.Context) error
	Upsert(context.Context, string, string, *agentsv1.PlaneInstallation, Tokens) (*agentsv1.PlaneInstallation, error)
	List(context.Context, string, string) ([]*agentsv1.PlaneInstallation, error)
	Get(context.Context, string, string) (*agentsv1.PlaneInstallation, error)
}
