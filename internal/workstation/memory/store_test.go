package memory_test

import (
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	"testing"
)

func TestContract(t *testing.T) {
	repotest.Run(t, func(*testing.T) ws.Repository { return memory.New() })
}
