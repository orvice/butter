package planeprocessing

import (
	"context"
	"errors"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"time"
)

type Claim struct {
	Owner     string
	Revision  int64
	Fence     uint64
	ExpiresAt time.Time
}
type ClaimRepository interface {
	Claim(context.Context, string, string, string, string, string, time.Time) (Claim, error)
	Renew(context.Context, string, Claim, time.Time) (Claim, error)
	Complete(context.Context, string, Claim) error
	ReclaimExpired(context.Context, string, string, string, string, string, time.Time) (Claim, error)
	RecoverExpired(context.Context, time.Time) error
}

var ErrNotFound = errors.New("plane processing record not found")
var ErrDuplicate = errors.New("plane processing activity already recorded")

var ErrClaimConflict = errors.New("plane processing claim conflict")
var ErrClaimExpired = errors.New("plane processing claim expired")

type Filter struct {
	WorkspaceID, AppID string
	Status             agentsv1.PlaneProcessingStatus
	Limit              int
}
type Repository interface {
	EnsureIndexes(context.Context) error
	Create(context.Context, *agentsv1.PlaneProcessingRecord) (*agentsv1.PlaneProcessingRecord, error)
	Get(context.Context, string, string) (*agentsv1.PlaneProcessingRecord, error)
	List(context.Context, Filter) ([]*agentsv1.PlaneProcessingRecord, error)
	UpdateStatus(context.Context, string, string, string, string, string, string, agentsv1.PlaneProcessingStatus, string) error
}

// RuntimeRepository is the full store used by the Plane worker: durable
// processing records plus the fenced execution claim.
type RuntimeRepository interface {
	Repository
	ClaimRepository
}
