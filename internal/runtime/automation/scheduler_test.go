package automation

import (
	"context"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestSchedulerRegistersEnabledScheduleAutomations(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{})
	enabled := scheduleAutomation("enabled", true)
	disabled := scheduleAutomation("disabled", false)
	manual := &agentsv1.Automation{
		Name:        "manual",
		Enabled:     true,
		WorkspaceId: "ws1",
		Trigger:     &agentsv1.AutomationTrigger{Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL},
	}
	for _, a := range []*agentsv1.Automation{enabled, disabled, manual} {
		if err := defRepo.Create(ctx, a); err != nil {
			t.Fatalf("Create(%s): %v", a.GetName(), err)
		}
	}

	scheduler, err := NewScheduler(ctx, defRepo, engine)
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	defer scheduler.Stop()
	if len(scheduler.entryID) != 1 {
		t.Fatalf("entry count = %d, want 1", len(scheduler.entryID))
	}
	if _, ok := scheduler.entryID[automationID("ws1", "enabled")]; !ok {
		t.Fatal("enabled schedule automation was not registered")
	}
}

func TestSchedulerRescheduleAndUnregister(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	scheduler, err := NewScheduler(ctx, defRepo, NewEngine(defRepo, NewMemoryRunRepo(), NewMemoryStepRunRepo(), EngineOptions{}))
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	defer scheduler.Stop()

	a := scheduleAutomation("daily", true)
	if err := scheduler.Reschedule(a); err != nil {
		t.Fatalf("Reschedule enabled: %v", err)
	}
	if len(scheduler.entryID) != 1 {
		t.Fatalf("entry count after enable = %d, want 1", len(scheduler.entryID))
	}
	a.Enabled = false
	if err := scheduler.Reschedule(a); err != nil {
		t.Fatalf("Reschedule disabled: %v", err)
	}
	if len(scheduler.entryID) != 0 {
		t.Fatalf("entry count after disable = %d, want 0", len(scheduler.entryID))
	}
}

func scheduleAutomation(name string, enabled bool) *agentsv1.Automation {
	return &agentsv1.Automation{
		Name:        name,
		Enabled:     enabled,
		WorkspaceId: "ws1",
		Trigger: &agentsv1.AutomationTrigger{
			Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_SCHEDULE,
			Schedule: &agentsv1.AutomationScheduleTrigger{
				Schedule: "@daily",
				Timezone: "UTC",
			},
		},
		Steps: []*agentsv1.AutomationStep{
			{Name: "noop", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK, CallWebhook: &agentsv1.AutomationCallWebhookStep{Url: "https://example.test"}},
		},
	}
}
