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
type TokenMaterial struct {
	WorkspaceID, AppID, InstallationID string
	AccessToken                        string
	ExpiresAt                          time.Time
	Revision                           int64
}

var ErrTokenCAS = errors.New("plane installation token revision conflict")

type Repository interface {
	EnsureIndexes(context.Context) error
	Upsert(context.Context, string, string, *agentsv1.PlaneInstallation, Tokens) (*agentsv1.PlaneInstallation, error)
	List(context.Context, string, string) ([]*agentsv1.PlaneInstallation, error)
	Get(context.Context, string, string) (*agentsv1.PlaneInstallation, error)
	GetRuntimeToken(context.Context, string, string, string) (TokenMaterial, error)
	CASRuntimeToken(context.Context, string, string, string, int64, Tokens) (TokenMaterial, error)
}
