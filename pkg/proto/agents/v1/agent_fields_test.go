package agentsv1_test

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestAgent_EnableOpenaiApiField(t *testing.T) {
	a := &agentsv1.Agent{
		Name:             "test-agent",
		EnableOpenaiApi:  true,
	}

	if !a.GetEnableOpenaiApi() {
		t.Error("expected EnableOpenaiApi to be true")
	}

	a2 := &agentsv1.Agent{Name: "other"}
	if a2.GetEnableOpenaiApi() {
		t.Error("expected EnableOpenaiApi to default to false")
	}
}
