package memory_test

import (
	"testing"

	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	repobindingmemory "go.orx.me/apps/butter/internal/repo/repobinding/memory"
	"go.orx.me/apps/butter/internal/repo/repobinding/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) repobindingrepo.Repository {
		return repobindingmemory.New()
	})
}
