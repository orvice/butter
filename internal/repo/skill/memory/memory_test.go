package memory_test

import (
	"testing"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	"go.orx.me/apps/butter/internal/repo/skill/repotest"
)

func TestMemoryRepositoryConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) skillrepo.Repository {
		return skillmemory.New()
	})
}
