package memory

import (
	"testing"

	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	"go.orx.me/apps/butter/internal/repo/memoryconfig/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(*testing.T) memoryconfigrepo.Repository { return New() })
}
