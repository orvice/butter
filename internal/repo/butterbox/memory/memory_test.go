package memory

import (
	"testing"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	"go.orx.me/apps/butter/internal/repo/butterbox/repotest"
)

func TestConformance(t *testing.T) {
	repotest.Run(t, func(*testing.T) butterboxrepo.Repository { return New() })
}
