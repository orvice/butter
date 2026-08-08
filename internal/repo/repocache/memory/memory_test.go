package memory_test

import (
	"testing"

	"go.orx.me/apps/butter/internal/repo/repocache"
	"go.orx.me/apps/butter/internal/repo/repocache/memory"
	"go.orx.me/apps/butter/internal/repo/repocache/repotest"
)

func TestRepositoryConformance(t *testing.T) {
	repotest.Run(t, func(*testing.T) repocache.Repository { return memory.New() })
}
