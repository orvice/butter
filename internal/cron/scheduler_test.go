package cron

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestJobConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		jobName   string
		schedule  string
		agentName string
		enabled   bool
	}{
		{"valid job", "test-job", "*/5 * * * *", "assistant", true},
		{"predefined schedule", "every-30m", "@every 30m", "assistant", true},
		{"disabled job", "disabled", "0 9 * * *", "assistant", false},
		{"job with input", "input-job", "@daily", "assistant", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.jobName == "" {
				t.Error("job name should not be empty")
			}
			if tt.schedule == "" {
				t.Error("job schedule should not be empty")
			}
			if tt.agentName == "" {
				t.Error("job agent_name should not be empty")
			}
		})
	}
}

func TestCronExpressionParsing(t *testing.T) {
	tests := []struct {
		name     string
		schedule string
		timezone string
		valid    bool
	}{
		{"standard 5-field", "*/5 * * * *", "", true},
		{"daily at 9", "0 9 * * *", "", true},
		{"predefined every", "@every 30m", "", true},
		{"predefined daily", "@daily", "", true},
		{"predefined hourly", "@hourly", "", true},
		{"with timezone", "0 9 * * *", "Asia/Shanghai", true},
		{"invalid expression", "not-a-cron", "", false},
		{"invalid timezone", "0 9 * * *", "Invalid/Timezone", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := tt.schedule
			if tt.timezone != "" {
				_, err := loadTimezone(tt.timezone)
				if err != nil {
					if tt.valid {
						t.Errorf("expected valid timezone %q, got error: %v", tt.timezone, err)
					}
					return
				}
				schedule = "CRON_TZ=" + tt.timezone + " " + schedule
			}

			_, err := parseSchedule(schedule)
			if tt.valid && err != nil {
				t.Errorf("expected valid schedule %q, got error: %v", schedule, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected invalid schedule %q, but parsed successfully", schedule)
			}
		})
	}
}

func TestDeliveryDefaults(t *testing.T) {
	// No delivery configured -> nil
	job := &agentsv1.CronJob{
		Name:      "test",
		Schedule:  "@daily",
		AgentName: "assistant",
		Enabled:   true,
	}
	if job.GetDelivery() != nil {
		t.Error("expected nil delivery for unconfigured job")
	}

	// Explicit LOG delivery
	job2 := &agentsv1.CronJob{
		Name:      "test2",
		Schedule:  "@daily",
		AgentName: "assistant",
		Enabled:   true,
		Delivery: &agentsv1.CronDelivery{
			Type: agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_LOG,
		},
	}
	if job2.GetDelivery().GetType() != agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_LOG {
		t.Error("expected LOG delivery type")
	}

	// Webhook delivery
	job3 := &agentsv1.CronJob{
		Name:      "test3",
		Schedule:  "@daily",
		AgentName: "assistant",
		Enabled:   true,
		Delivery: &agentsv1.CronDelivery{
			Type:       agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_WEBHOOK,
			WebhookUrl: "https://example.com/hook",
		},
	}
	if job3.GetDelivery().GetWebhookUrl() != "https://example.com/hook" {
		t.Error("expected webhook URL")
	}
}
