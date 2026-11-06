package memory

import (
	"testing"

	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	"go.orx.me/apps/butter/internal/repo/plane/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(*testing.T) planerepo.Repository { return New() })
}
