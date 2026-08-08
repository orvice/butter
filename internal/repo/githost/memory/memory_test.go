package memory_test

import (
	"testing"

	githostrepo "go.orx.me/apps/butter/internal/repo/githost"
	githostmemory "go.orx.me/apps/butter/internal/repo/githost/memory"
	"go.orx.me/apps/butter/internal/repo/githost/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) githostrepo.Repository {
		return githostmemory.New()
	})
}
