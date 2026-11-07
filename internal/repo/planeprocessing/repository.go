package planeprocessing

import (
	"context"
	"errors"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var ErrNotFound = errors.New("plane processing record not found")
var ErrDuplicate = errors.New("plane processing activity already recorded")

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
}
