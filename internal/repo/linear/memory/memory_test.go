package memory

import (
	"testing"

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	"go.orx.me/apps/butter/internal/repo/linear/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(*testing.T) linearrepo.Repository { return New() })
}
