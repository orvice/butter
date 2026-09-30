package memory

import (
	"testing"

	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	"go.orx.me/apps/butter/internal/repo/linearprocessing/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(*testing.T) linearprocessing.Repository { return New() })
}
