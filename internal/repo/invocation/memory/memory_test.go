package memory

import (
	"testing"

	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/repo/invocation/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(*testing.T) repotest.Open {
		store := New()
		return func(owner string) invocation.Repository { return store.WithOwner(owner) }
	})
}
