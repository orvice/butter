package a2uitool

import (
	"testing"

	"go.orx.me/apps/butter/internal/testsupport/tooltest"
)

func TestToolParametersAreDescribed(t *testing.T) {
	tooltest.RequireParamDescriptions(t, NewToolset().tool)
}
