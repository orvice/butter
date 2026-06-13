package cron

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/notify"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

type testNotifyGroupRepo struct {
	group *agentsv1.NotifyGroup
}

func (r *testNotifyGroupRepo) ListNotifyGroups(context.Context, string) ([]*agentsv1.NotifyGroup, error) {
	return []*agentsv1.NotifyGroup{r.group}, nil
}

func (r *testNotifyGroupRepo) GetNotifyGroup(context.Context, string, string) (*agentsv1.NotifyGroup, error) {
	if r.group == nil {
		return nil, configrepo.ErrNotFound
	}
	return r.group, nil
}

func (r *testNotifyGroupRepo) CreateNotifyGroup(context.Context, string, *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	return nil, errors.New("not implemented")
}

func (r *testNotifyGroupRepo) UpdateNotifyGroup(context.Context, string, *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	return nil, errors.New("not implemented")
}

func (r *testNotifyGroupRepo) DeleteNotifyGroup(context.Context, string, string) error {
	return errors.New("not implemented")
}

func (r *testNotifyGroupRepo) ListNotifyGroupsAcrossWorkspaces(context.Context) ([]*agentsv1.NotifyGroup, error) {
	return []*agentsv1.NotifyGroup{r.group}, nil
}

type testChannelRepo struct {
	channel *agentsv1.AgentChannel
}

func (r *testChannelRepo) ListChannels(context.Context, string) ([]*agentsv1.AgentChannel, error) {
	return []*agentsv1.AgentChannel{r.channel}, nil
}

func (r *testChannelRepo) GetChannel(_ context.Context, workspaceID, name string) (*agentsv1.AgentChannel, error) {
	if r.channel == nil || r.channel.GetWorkspaceId() != workspaceID || r.channel.GetName() != name {
		return nil, configrepo.ErrNotFound
	}
	return r.channel, nil
}

func (r *testChannelRepo) CreateChannel(context.Context, string, *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return nil, errors.New("not implemented")
}

func (r *testChannelRepo) UpdateChannel(context.Context, string, *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return nil, errors.New("not implemented")
}

func (r *testChannelRepo) DeleteChannel(context.Context, string, string) error {
	return errors.New("not implemented")
}

func (r *testChannelRepo) ListChannelsAcrossWorkspaces(context.Context) ([]*agentsv1.AgentChannel, error) {
	return []*agentsv1.AgentChannel{r.channel}, nil
}

type testChannelSender struct {
	channel *agentsv1.AgentChannel
	chatID  string
	text    string
}

func (s *testChannelSender) Send(_ context.Context, channel *agentsv1.AgentChannel, chatID, text string) error {
	s.channel = channel
	s.chatID = chatID
	s.text = text
	return nil
}

type testCronRunner struct {
	output      string
	errs        []error
	block       chan struct{}
	runSSECalls int
}

func (r *testCronRunner) HasAgentInWorkspace(string, string) bool {
	return true
}

func (r *testCronRunner) RunSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.runSSECalls++
	if r.block != nil {
		select {
		case <-r.block:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if len(r.errs) > 0 {
		err := r.errs[0]
		r.errs = r.errs[1:]
		if err != nil {
			return "", err
		}
	}
	return r.output, nil
}

type testExecutionRepo struct {
	saved *agentsv1.CronExecution
}

func (r *testExecutionRepo) Save(_ context.Context, exec *agentsv1.CronExecution) error {
	r.saved = exec
	return nil
}

func (r *testExecutionRepo) List(context.Context, string, string, int32, string) ([]*agentsv1.CronExecution, string, error) {
	return nil, "", errors.New("not implemented")
}

func (r *testExecutionRepo) GetByID(context.Context, string) (*agentsv1.CronExecution, error) {
	return nil, errors.New("not implemented")
}

func (r *testExecutionRepo) ListByTimeRange(context.Context, string, string, time.Time, time.Time) ([]*agentsv1.CronExecution, error) {
	return nil, errors.New("not implemented")
}

func TestExecuteJobUsesSSERunner(t *testing.T) {
	cronRunner := &testCronRunner{output: "done"}
	execRepo := &testExecutionRepo{}
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   cronRunner,
		execRepo: execRepo,
		notifier: notify.NewSender(nil),
	}

	exec := s.executeJob(&agentsv1.CronJob{
		Name:        "daily-summary",
		WorkspaceId: "ws1",
		AgentName:   "assistant",
		Input:       "summarize",
	})

	if cronRunner.runSSECalls != 1 {
		t.Fatalf("RunSSE calls = %d, want 1", cronRunner.runSSECalls)
	}
	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS {
		t.Fatalf("status = %s, want success", exec.GetStatus())
	}
	if exec.GetOutput() != "done" {
		t.Fatalf("output = %q, want %q", exec.GetOutput(), "done")
	}
	if execRepo.saved != exec {
		t.Fatal("execution was not saved")
	}
}

func TestLegacyCronDefaultsPreservePreviousBehavior(t *testing.T) {
	cronRunner := &testCronRunner{output: "done"}
	execRepo := &testExecutionRepo{}
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   cronRunner,
		execRepo: execRepo,
		notifier: notify.NewSender(nil),
	}

	exec := s.executeJob(&agentsv1.CronJob{
		Name:        "legacy",
		WorkspaceId: "ws1",
		AgentName:   "assistant",
		Schedule:    "@daily",
	})

	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS {
		t.Fatalf("status = %s, want success", exec.GetStatus())
	}
	if exec.GetInput() != "execute" {
		t.Fatalf("input = %q, want execute", exec.GetInput())
	}
	if exec.GetAttemptCount() != 1 {
		t.Fatalf("attempt_count = %d, want 1", exec.GetAttemptCount())
	}
	if exec.GetTruncated() {
		t.Fatal("legacy default should not truncate output")
	}
	if cronRunner.runSSECalls != 1 {
		t.Fatalf("RunSSE calls = %d, want 1", cronRunner.runSSECalls)
	}
	if execRepo.saved != exec {
		t.Fatal("execution was not saved")
	}
}

func TestExecuteJobRetriesUntilSuccess(t *testing.T) {
	cronRunner := &testCronRunner{output: "done", errs: []error{errors.New("temporary failure"), nil}}
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   cronRunner,
		execRepo: &testExecutionRepo{},
		notifier: notify.NewSender(nil),
	}

	exec := s.executeJob(&agentsv1.CronJob{
		Name:        "retry-job",
		WorkspaceId: "ws1",
		AgentName:   "assistant",
		Schedule:    "@daily",
		Retry:       &agentsv1.CronRetryPolicy{MaxAttempts: 1},
	})

	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS {
		t.Fatalf("status = %s, want success", exec.GetStatus())
	}
	if exec.GetAttemptCount() != 2 {
		t.Fatalf("attempt_count = %d, want 2", exec.GetAttemptCount())
	}
	if cronRunner.runSSECalls != 2 {
		t.Fatalf("RunSSE calls = %d, want 2", cronRunner.runSSECalls)
	}
}

func TestExecuteJobTimeoutCancelsInvocation(t *testing.T) {
	cronRunner := &testCronRunner{block: make(chan struct{})}
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   cronRunner,
		execRepo: &testExecutionRepo{},
		notifier: notify.NewSender(nil),
	}

	exec := s.executeJob(&agentsv1.CronJob{
		Name:        "timeout-job",
		WorkspaceId: "ws1",
		AgentName:   "assistant",
		Schedule:    "@daily",
		Timeout:     durationpb.New(10 * time.Millisecond),
	})

	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR {
		t.Fatalf("status = %s, want error", exec.GetStatus())
	}
	if exec.GetError() == "" {
		t.Fatal("expected timeout error text")
	}
	if exec.GetAttemptCount() != 1 {
		t.Fatalf("attempt_count = %d, want 1", exec.GetAttemptCount())
	}
}

func TestExecuteJobConcurrencySkipRecordsSkipped(t *testing.T) {
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   &testCronRunner{output: "done"},
		execRepo: &testExecutionRepo{},
		notifier: notify.NewSender(nil),
		running: map[string]*runningJob{
			jobKey("ws1", "skip-job"): {cancel: func() {}, done: make(chan struct{})},
		},
	}

	exec := s.executeJobWithTrigger(&agentsv1.CronJob{
		Name:        "skip-job",
		WorkspaceId: "ws1",
		AgentName:   "assistant",
		Schedule:    "@daily",
	}, agentsv1.CronExecutionTriggerType_CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE)

	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SKIPPED {
		t.Fatalf("status = %s, want skipped", exec.GetStatus())
	}
	if exec.GetSkippedReason() == "" {
		t.Fatal("skipped reason should be recorded")
	}
}

func TestExecuteJobConcurrencyAllowStartsOverlapping(t *testing.T) {
	cronRunner := &testCronRunner{output: "done"}
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   cronRunner,
		execRepo: &testExecutionRepo{},
		notifier: notify.NewSender(nil),
		running: map[string]*runningJob{
			jobKey("ws1", "allow-job"): {cancel: func() {}, done: make(chan struct{})},
		},
	}

	exec := s.executeJobWithTrigger(&agentsv1.CronJob{
		Name:              "allow-job",
		WorkspaceId:       "ws1",
		AgentName:         "assistant",
		Schedule:          "@daily",
		ConcurrencyPolicy: agentsv1.CronConcurrencyPolicy_CRON_CONCURRENCY_POLICY_ALLOW,
	}, agentsv1.CronExecutionTriggerType_CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE)

	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS {
		t.Fatalf("status = %s, want success", exec.GetStatus())
	}
	if cronRunner.runSSECalls != 1 {
		t.Fatalf("RunSSE calls = %d, want 1", cronRunner.runSSECalls)
	}
}

func TestShouldDeliverNotifyPolicy(t *testing.T) {
	success := &agentsv1.CronExecution{Status: agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS}
	failure := &agentsv1.CronExecution{Status: agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR}

	if shouldDeliver(&agentsv1.CronJob{NotifyOn: agentsv1.CronNotifyOn_CRON_NOTIFY_ON_FAILURE}, success) {
		t.Fatal("failure-only policy should not deliver successful execution")
	}
	if !shouldDeliver(&agentsv1.CronJob{NotifyOn: agentsv1.CronNotifyOn_CRON_NOTIFY_ON_FAILURE}, failure) {
		t.Fatal("failure-only policy should deliver failed execution")
	}
	if shouldDeliver(&agentsv1.CronJob{NotifyOn: agentsv1.CronNotifyOn_CRON_NOTIFY_ON_SUCCESS}, failure) {
		t.Fatal("success-only policy should not deliver failed execution")
	}
}

type fanoutTransport struct {
	mu      sync.Mutex
	paths   []string
	hungHit chan struct{}
}

func (t *fanoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.paths = append(t.paths, req.URL.Path)
	t.mu.Unlock()
	if req.URL.Path == "/hang" {
		select {
		case <-t.hungHit:
		default:
			close(t.hungHit)
		}
		<-req.Context().Done()
		return nil, req.Context().Err()
	}
	if req.URL.Path == "/fail" {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func (t *fanoutTransport) saw(path string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, got := range t.paths {
		if got == path {
			return true
		}
	}
	return false
}

func TestDeliverNotifyGroupTimeoutFailureContinuesFanout(t *testing.T) {
	origTimeout := notifyTargetTimeout
	notifyTargetTimeout = 25 * time.Millisecond
	defer func() { notifyTargetTimeout = origTimeout }()

	transport := &fanoutTransport{hungHit: make(chan struct{})}
	s := &Scheduler{
		ctx: context.Background(),
		groupRepo: &testNotifyGroupRepo{group: &agentsv1.NotifyGroup{
			Name:    "ops",
			Enabled: true,
			Targets: []*agentsv1.NotifyTarget{
				{
					Name:    "hang",
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
					Lark:    &agentsv1.LarkNotifyTarget{WebhookUrl: "https://notify.test/hang"},
				},
				{
					Name:    "fail",
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK,
					Discord: &agentsv1.DiscordNotifyTarget{
						WebhookUrl: "https://notify.test/fail",
					},
				},
				{
					Name:    "ok",
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK,
					Discord: &agentsv1.DiscordNotifyTarget{
						WebhookUrl: "https://notify.test/ok",
					},
				},
			},
		}},
		notifier: notify.NewSender(&http.Client{Transport: transport}),
	}

	done := make(chan struct{})
	go func() {
		s.deliverNotifyGroup(&agentsv1.CronJob{
			Name:        "job1",
			WorkspaceId: "ws1",
			Delivery: &agentsv1.CronDelivery{
				Type:            agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_NOTIFY_GROUP,
				NotifyGroupName: "ops",
			},
		}, &agentsv1.CronExecution{
			JobName:   "job1",
			Status:    agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS,
			Output:    "done",
			StartedAt: timestamppb.Now(),
		})
		close(done)
	}()

	select {
	case <-transport.hungHit:
	case <-time.After(time.Second):
		t.Fatal("hung target was not called")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("notify group delivery did not return after target timeout")
	}
	if !transport.saw("/fail") {
		t.Fatal("expected fan-out to continue to failing target after timeout")
	}
	if !transport.saw("/ok") {
		t.Fatal("expected fan-out to continue to later successful target after failure")
	}
}

func TestDeliverChannelSendsThroughConfiguredAgentChannel(t *testing.T) {
	sender := &testChannelSender{}
	s := &Scheduler{
		ctx: context.Background(),
		channelRepo: &testChannelRepo{channel: &agentsv1.AgentChannel{
			Name:        "ops-chat",
			WorkspaceId: "ws1",
			Enabled:     true,
			Platform:    agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD,
		}},
		channelSender: sender,
	}

	s.deliverChannel(&agentsv1.CronJob{
		Name:        "job1",
		WorkspaceId: "ws1",
		Delivery: &agentsv1.CronDelivery{
			Type:        agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_CHANNEL,
			ChannelName: "ops-chat",
			ChatId:      "chan-123",
		},
	}, &agentsv1.CronExecution{
		Id:      "exec-1",
		JobName: "job1",
		Status:  agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS,
		Output:  "done",
	})

	if sender.channel == nil || sender.channel.GetName() != "ops-chat" {
		t.Fatalf("sender channel = %v, want ops-chat", sender.channel)
	}
	if sender.chatID != "chan-123" {
		t.Fatalf("chatID = %q, want chan-123", sender.chatID)
	}
	if sender.text == "" {
		t.Fatal("expected non-empty delivery message")
	}
}
