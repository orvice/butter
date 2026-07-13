package cron

import (
	"context"
	"errors"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
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
	output   string
	pending  []runner.PendingInput
	errs     []error
	block    chan struct{}
	runCalls int
	// results, when non-empty, is consumed one entry per call instead of
	// output/pending — for tests where consecutive runs behave differently.
	results []*runner.TurnResult
	// ctxInfos records the ContextInfo of every call, in order.
	ctxInfos []*agentsv1.ContextInfo
	// onTurn mimics runner.Service's turn-listener dispatch: called after
	// every turn with the same arguments a registered TurnListener receives.
	onTurn func(*agentsv1.ContextInfo, *runner.TurnResult, error)
}

func (r *testCronRunner) HasAgentInWorkspace(string, string) bool {
	return true
}

func (r *testCronRunner) RunTurnSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, ctxInfo *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (*runner.TurnResult, error) {
	r.runCalls++
	r.ctxInfos = append(r.ctxInfos, ctxInfo)
	if r.block != nil {
		select {
		case <-r.block:
		case <-ctx.Done():
			return &runner.TurnResult{}, ctx.Err()
		}
	}
	if len(r.errs) > 0 {
		err := r.errs[0]
		r.errs = r.errs[1:]
		if err != nil {
			return &runner.TurnResult{}, err
		}
	}
	turn := &runner.TurnResult{Output: r.output, Pending: r.pending}
	if len(r.results) > 0 {
		turn = r.results[0]
		r.results = r.results[1:]
	}
	if r.onTurn != nil {
		r.onTurn(ctxInfo, turn, nil)
	}
	return turn, nil
}

type testExecutionRepo struct {
	saved *agentsv1.CronExecution
	execs map[string]*agentsv1.CronExecution
}

func (r *testExecutionRepo) Save(_ context.Context, exec *agentsv1.CronExecution) error {
	r.saved = exec
	if r.execs == nil {
		r.execs = map[string]*agentsv1.CronExecution{}
	}
	r.execs[exec.GetId()] = exec
	return nil
}

func (r *testExecutionRepo) ListWaitingBySessionAcrossWorkspaces(_ context.Context, appName, userID, sessionID string) ([]*agentsv1.CronExecution, error) {
	var out []*agentsv1.CronExecution
	for _, e := range r.execs {
		if e.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT &&
			e.GetSessionAppName() == appName &&
			e.GetSessionUserId() == userID &&
			e.GetSessionId() == sessionID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].GetStartedAt().AsTime().Before(out[j].GetStartedAt().AsTime())
	})
	return out, nil
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

func TestExecuteJobPausedWorkflowRecordsWaitingInput(t *testing.T) {
	cronRunner := &testCronRunner{
		output:  "Deploy v2 to production?",
		pending: []runner.PendingInput{{InterruptID: "int-1", Question: "Deploy v2 to production?"}},
	}
	execRepo := &testExecutionRepo{}
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   cronRunner,
		execRepo: execRepo,
		notifier: notify.NewSender(nil),
	}

	exec := s.executeJob(&agentsv1.CronJob{
		Name:        "approve-deploy",
		WorkspaceId: "ws1",
		AgentName:   "release-flow",
		Schedule:    "@daily",
	})

	if exec.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
		t.Fatalf("status = %s, want waiting_input", exec.GetStatus())
	}
	if exec.GetOutput() != "Deploy v2 to production?" {
		t.Fatalf("output = %q, want the node's question", exec.GetOutput())
	}
	if exec.GetSessionAppName() != "cron:approve-deploy" {
		t.Fatalf("session_app_name = %q, want cron:approve-deploy", exec.GetSessionAppName())
	}
	if exec.GetSessionUserId() != "cron:approve-deploy" {
		t.Fatalf("session_user_id = %q, want cron:approve-deploy", exec.GetSessionUserId())
	}
	// The session is per-execution (issue #131): reruns of the job must not
	// share the session where this execution waits for its human answer.
	if want := "cron:approve-deploy:" + exec.GetId(); exec.GetSessionId() != want {
		t.Fatalf("session_id = %q, want %q", exec.GetSessionId(), want)
	}
	if exec.GetFinishedAt() != nil {
		t.Fatalf("finished_at = %v, want unset: a waiting execution has not finished", exec.GetFinishedAt())
	}
	if execRepo.saved != exec {
		t.Fatal("waiting execution was not persisted")
	}
}

func TestPausedWorkflowDeliversQuestionWithSessionCoordinates(t *testing.T) {
	sender := &testChannelSender{}
	s := &Scheduler{
		ctx: context.Background(),
		runner: &testCronRunner{
			output:  "Deploy v2 to production?",
			pending: []runner.PendingInput{{InterruptID: "int-1", Question: "Deploy v2 to production?"}},
		},
		execRepo: &testExecutionRepo{},
		notifier: notify.NewSender(nil),
		channelRepo: &testChannelRepo{channel: &agentsv1.AgentChannel{
			Name:        "ops-chat",
			WorkspaceId: "ws1",
			Enabled:     true,
			Platform:    agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD,
		}},
		channelSender: sender,
	}

	exec := s.executeJob(&agentsv1.CronJob{
		Name:        "approve-deploy",
		WorkspaceId: "ws1",
		AgentName:   "release-flow",
		Schedule:    "@daily",
		Delivery: &agentsv1.CronDelivery{
			Type:        agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_CHANNEL,
			ChannelName: "ops-chat",
			ChatId:      "chan-123",
		},
	})

	if sender.text == "" {
		t.Fatal("expected a delivery for the paused execution")
	}
	for _, want := range []string{
		"Deploy v2 to production?", // the node's question
		"waiting_input",            // the state, so the reader knows it blocks
		"agent_name=release-flow",  // ReplySession coordinates follow
		"app_name=cron:approve-deploy",
		"user_id=cron:approve-deploy",
		// The full per-execution session ID, not just the job scope — a reply
		// to a truncated ID would land on the wrong session.
		"session_id=cron:approve-deploy:" + exec.GetId(),
	} {
		if !strings.Contains(sender.text, want) {
			t.Fatalf("delivery message missing %q:\n%s", want, sender.text)
		}
	}
}

func TestDeliverWebhookWaitingInputCarriesSessionCoordinates(t *testing.T) {
	var payload map[string]string
	received := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		close(received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	s := &Scheduler{ctx: context.Background()}
	s.deliverWebhook(&agentsv1.CronJob{
		Name:        "approve-deploy",
		WorkspaceId: "ws1",
		Delivery: &agentsv1.CronDelivery{
			Type:       agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_WEBHOOK,
			WebhookUrl: srv.URL,
		},
	}, &agentsv1.CronExecution{
		Id:             "exec-1",
		JobName:        "approve-deploy",
		AgentName:      "release-flow",
		Status:         agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT,
		Output:         "Deploy v2 to production?",
		StartedAt:      timestamppb.Now(),
		SessionAppName: "cron:approve-deploy",
		SessionUserId:  "cron:approve-deploy",
		SessionId:      "cron:approve-deploy:exec-1",
	})

	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("webhook was not called")
	}
	if payload["status"] != "waiting_input" {
		t.Fatalf("status = %q, want waiting_input", payload["status"])
	}
	if payload["output"] != "Deploy v2 to production?" {
		t.Fatalf("output = %q, want the question", payload["output"])
	}
	for key, want := range map[string]string{
		"agent_name":       "release-flow",
		"session_app_name": "cron:approve-deploy",
		"session_user_id":  "cron:approve-deploy",
		"session_id":       "cron:approve-deploy:exec-1",
	} {
		if payload[key] != want {
			t.Fatalf("payload[%q] = %q, want %q", key, payload[key], want)
		}
	}
}

func TestDeliverNotifyGroupWaitingInputCarriesSessionCoordinates(t *testing.T) {
	var got notify.Message
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		var lark struct {
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		_ = json.Unmarshal(body, &lark)
		got.Text = lark.Content.Text
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: make(http.Header), Request: req}, nil
	})

	s := &Scheduler{
		ctx: context.Background(),
		groupRepo: &testNotifyGroupRepo{group: &agentsv1.NotifyGroup{
			Name:    "ops",
			Enabled: true,
			Targets: []*agentsv1.NotifyTarget{{
				Name:    "lark",
				Enabled: true,
				Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
				Lark:    &agentsv1.LarkNotifyTarget{WebhookUrl: "https://notify.test/lark"},
			}},
		}},
		notifier: notify.NewSender(&http.Client{Transport: transport}),
	}

	s.deliverNotifyGroup(&agentsv1.CronJob{
		Name:        "approve-deploy",
		WorkspaceId: "ws1",
		Delivery: &agentsv1.CronDelivery{
			Type:            agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_NOTIFY_GROUP,
			NotifyGroupName: "ops",
		},
	}, &agentsv1.CronExecution{
		Id:             "exec-1",
		JobName:        "approve-deploy",
		AgentName:      "release-flow",
		Status:         agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT,
		Output:         "Deploy v2 to production?",
		StartedAt:      timestamppb.Now(),
		SessionAppName: "cron:approve-deploy",
		SessionUserId:  "cron:approve-deploy",
		SessionId:      "cron:approve-deploy:exec-1",
	})

	for _, want := range []string{
		"Deploy v2 to production?",
		"agent_name=release-flow",
		"app_name=cron:approve-deploy",
		"user_id=cron:approve-deploy",
		"session_id=cron:approve-deploy:exec-1",
	} {
		if !strings.Contains(got.Text, want) {
			t.Fatalf("notify text missing %q:\n%s", want, got.Text)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestShouldDeliverWaitingInputIgnoresNotifyPolicy(t *testing.T) {
	waiting := &agentsv1.CronExecution{Status: agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT}

	// The question must reach a human whatever the notify_on policy says —
	// otherwise the workflow waits forever with nobody told.
	for _, policy := range []agentsv1.CronNotifyOn{
		agentsv1.CronNotifyOn_CRON_NOTIFY_ON_FAILURE,
		agentsv1.CronNotifyOn_CRON_NOTIFY_ON_SUCCESS,
		agentsv1.CronNotifyOn_CRON_NOTIFY_ON_ALWAYS,
	} {
		if !shouldDeliver(&agentsv1.CronJob{NotifyOn: policy}, waiting) {
			t.Fatalf("policy %s should not suppress a waiting_input delivery", policy)
		}
	}
}

type testJobRepo struct {
	job *agentsv1.CronJob
}

func (r *testJobRepo) List(context.Context, string) ([]*agentsv1.CronJob, error) {
	return []*agentsv1.CronJob{r.job}, nil
}

func (r *testJobRepo) ListAll(context.Context) ([]*agentsv1.CronJob, error) {
	return []*agentsv1.CronJob{r.job}, nil
}

func (r *testJobRepo) Get(_ context.Context, workspaceID, name string) (*agentsv1.CronJob, error) {
	if r.job == nil || r.job.GetWorkspaceId() != workspaceID || r.job.GetName() != name {
		return nil, errors.New("job not found")
	}
	return r.job, nil
}

func (r *testJobRepo) Create(context.Context, *agentsv1.CronJob) error {
	return errors.New("not implemented")
}

func (r *testJobRepo) Update(context.Context, *agentsv1.CronJob) error {
	return errors.New("not implemented")
}

func (r *testJobRepo) Delete(context.Context, string, string) error {
	return errors.New("not implemented")
}

func waitingExecFixture() *agentsv1.CronExecution {
	return &agentsv1.CronExecution{
		Id:             "exec-1",
		JobName:        "approve-deploy",
		AgentName:      "release-flow",
		Status:         agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT,
		Input:          "execute",
		Output:         "Deploy v2 to production?",
		StartedAt:      timestamppb.New(time.Now().Add(-time.Minute)),
		SessionAppName: "cron:approve-deploy",
		SessionUserId:  "cron:approve-deploy",
		SessionId:      "cron:approve-deploy:exec-1",
		WorkspaceId:    "ws1",
	}
}

func cronSessionCtxInfo() *agentsv1.ContextInfo {
	return &agentsv1.ContextInfo{
		ChannelName: "cron:approve-deploy",
		UserId:      "cron:approve-deploy",
		SessionId:   "cron:approve-deploy:exec-1",
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
	}
}

// TestHandleSessionDeletedCancelsWaitingExecution: issue #132 — deleting a
// paused execution's session is the documented way to abandon the workflow
// (ADR 0002); the WAITING_INPUT record must reach CANCELLED with a reason and
// the cancellation delivered through the job's target, instead of waiting
// forever on a session that no longer exists.
func TestHandleSessionDeletedCancelsWaitingExecution(t *testing.T) {
	sender := &testChannelSender{}
	execRepo := &testExecutionRepo{}
	if err := execRepo.Save(context.Background(), waitingExecFixture()); err != nil {
		t.Fatal(err)
	}
	s := &Scheduler{
		ctx:      context.Background(),
		execRepo: execRepo,
		jobRepo: &testJobRepo{job: &agentsv1.CronJob{
			Name:        "approve-deploy",
			WorkspaceId: "ws1",
			AgentName:   "release-flow",
			Schedule:    "@daily",
			Delivery: &agentsv1.CronDelivery{
				Type:        agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_CHANNEL,
				ChannelName: "ops-chat",
				ChatId:      "chan-123",
			},
		}},
		channelRepo: &testChannelRepo{channel: &agentsv1.AgentChannel{
			Name:        "ops-chat",
			WorkspaceId: "ws1",
			Enabled:     true,
			Platform:    agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD,
		}},
		channelSender: sender,
		notifier:      notify.NewSender(nil),
	}

	s.HandleSessionDeleted("cron:approve-deploy", "cron:approve-deploy", "cron:approve-deploy:exec-1")

	got := execRepo.execs["exec-1"]
	if got.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_CANCELLED {
		t.Fatalf("status = %s, want cancelled", got.GetStatus())
	}
	if got.GetError() == "" || !strings.Contains(got.GetError(), "session deleted") {
		t.Fatalf("error = %q, want a session-deleted reason", got.GetError())
	}
	if got.GetFinishedAt() == nil {
		t.Fatal("finished_at should be set on cancellation")
	}
	if got.GetDurationMs() <= 0 {
		t.Fatalf("duration_ms = %d, want the wall time since the run started", got.GetDurationMs())
	}
	if !strings.Contains(sender.text, "cancelled") {
		t.Fatalf("cancellation was not delivered: %q", sender.text)
	}
}

// TestHandleSessionDeletedIgnoresUnrelatedSessions: deleting a session that
// is not the waiting execution's — another run's session, another job's, or
// ordinary chat traffic — leaves every execution record untouched.
func TestHandleSessionDeletedIgnoresUnrelatedSessions(t *testing.T) {
	cases := []struct {
		name      string
		appName   string
		userID    string
		sessionID string
	}{
		{
			name:      "another execution of the same job",
			appName:   "cron:approve-deploy",
			userID:    "cron:approve-deploy",
			sessionID: "cron:approve-deploy:exec-2",
		},
		{
			name:      "another job's session",
			appName:   "cron:other-job",
			userID:    "cron:other-job",
			sessionID: "cron:other-job:exec-1",
		},
		{
			name:      "non-cron chat session",
			appName:   "telegram",
			userID:    "u1",
			sessionID: "chat:100",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sender := &testChannelSender{}
			execRepo := &testExecutionRepo{}
			if err := execRepo.Save(context.Background(), waitingExecFixture()); err != nil {
				t.Fatal(err)
			}
			s := &Scheduler{
				ctx:           context.Background(),
				execRepo:      execRepo,
				jobRepo:       &testJobRepo{},
				channelSender: sender,
				notifier:      notify.NewSender(nil),
			}

			s.HandleSessionDeleted(tc.appName, tc.userID, tc.sessionID)

			got := execRepo.execs["exec-1"]
			if got.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
				t.Fatalf("status = %s, want still waiting_input", got.GetStatus())
			}
			if sender.text != "" {
				t.Fatalf("unexpected delivery: %q", sender.text)
			}
		})
	}
}

// TestHandleTurnCompletesWaitingExecution: a reply on the cron session (via
// ReplySession) resumes the workflow; when the resumed turn ends with no
// pending Interrupt, the waiting execution reaches its terminal state and
// the final output is delivered through the job's configured target.
func TestHandleTurnCompletesWaitingExecution(t *testing.T) {
	sender := &testChannelSender{}
	execRepo := &testExecutionRepo{}
	if err := execRepo.Save(context.Background(), waitingExecFixture()); err != nil {
		t.Fatal(err)
	}
	s := &Scheduler{
		ctx:      context.Background(),
		execRepo: execRepo,
		jobRepo: &testJobRepo{job: &agentsv1.CronJob{
			Name:        "approve-deploy",
			WorkspaceId: "ws1",
			AgentName:   "release-flow",
			Schedule:    "@daily",
			Delivery: &agentsv1.CronDelivery{
				Type:        agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_CHANNEL,
				ChannelName: "ops-chat",
				ChatId:      "chan-123",
			},
		}},
		channelRepo: &testChannelRepo{channel: &agentsv1.AgentChannel{
			Name:        "ops-chat",
			WorkspaceId: "ws1",
			Enabled:     true,
			Platform:    agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD,
		}},
		channelSender: sender,
		notifier:      notify.NewSender(nil),
	}

	s.HandleTurn(cronSessionCtxInfo(), &runner.TurnResult{Output: "deployed v2"}, nil)

	got := execRepo.execs["exec-1"]
	if got.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS {
		t.Fatalf("status = %s, want success", got.GetStatus())
	}
	if got.GetOutput() != "deployed v2" {
		t.Fatalf("output = %q, want the resumed turn's output", got.GetOutput())
	}
	if got.GetFinishedAt() == nil {
		t.Fatal("finished_at should be set on completion")
	}
	if got.GetDurationMs() <= 0 {
		t.Fatalf("duration_ms = %d, want the wall time since the run started", got.GetDurationMs())
	}
	if !strings.Contains(sender.text, "deployed v2") || !strings.Contains(sender.text, "success") {
		t.Fatalf("final result was not delivered: %q", sender.text)
	}
}

// TestHandleTurnIgnoresUnfinishedAndUnrelatedTurns: the execution stays
// WAITING_INPUT while the workflow is still paused, when the resume turn
// errored, and turns on unrelated sessions never touch it.
func TestHandleTurnIgnoresUnfinishedAndUnrelatedTurns(t *testing.T) {
	cases := []struct {
		name    string
		ctxInfo *agentsv1.ContextInfo
		turn    *runner.TurnResult
		err     error
	}{
		{
			name:    "still interrupted",
			ctxInfo: cronSessionCtxInfo(),
			turn: &runner.TurnResult{
				Output:  "Second question?",
				Pending: []runner.PendingInput{{InterruptID: "int-2", Question: "Second question?"}},
			},
		},
		{
			name:    "turn errored",
			ctxInfo: cronSessionCtxInfo(),
			turn:    &runner.TurnResult{},
			err:     errors.New("model unavailable"),
		},
		{
			name: "unrelated session",
			ctxInfo: &agentsv1.ContextInfo{
				ChannelName: "telegram",
				UserId:      "u1",
				SessionId:   "chat:100",
			},
			turn: &runner.TurnResult{Output: "hello"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sender := &testChannelSender{}
			execRepo := &testExecutionRepo{}
			if err := execRepo.Save(context.Background(), waitingExecFixture()); err != nil {
				t.Fatal(err)
			}
			s := &Scheduler{
				ctx:           context.Background(),
				execRepo:      execRepo,
				jobRepo:       &testJobRepo{},
				channelSender: sender,
				notifier:      notify.NewSender(nil),
			}

			s.HandleTurn(tc.ctxInfo, tc.turn, tc.err)

			got := execRepo.execs["exec-1"]
			if got.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
				t.Fatalf("status = %s, want still waiting_input", got.GetStatus())
			}
			if sender.text != "" {
				t.Fatalf("unexpected delivery: %q", sender.text)
			}
		})
	}
}

// TestScheduledRerunWhileExecutionWaitsDoesNotAnswerIt: issue #131 — while
// an execution sits in WAITING_INPUT, the job keeps firing on schedule (the
// concurrency slot is released when the pausing run returns). The rescheduled
// run must not post its input onto the waiting execution's session: per ADR
// 0002 the runner would treat that input as the human's answer to the pending
// Interrupt, and the rerun's completion would close the older WAITING_INPUT
// record with unrelated output.
func TestScheduledRerunWhileExecutionWaitsDoesNotAnswerIt(t *testing.T) {
	job := &agentsv1.CronJob{
		Name:        "approve-deploy",
		WorkspaceId: "ws1",
		AgentName:   "release-flow",
		Schedule:    "@daily",
	}
	execRepo := &testExecutionRepo{}
	cronRunner := &testCronRunner{
		results: []*runner.TurnResult{
			{
				Output:  "Deploy v2 to production?",
				Pending: []runner.PendingInput{{InterruptID: "int-1", Question: "Deploy v2 to production?"}},
			},
			{Output: "unrelated next-day output"},
		},
	}
	s := &Scheduler{
		ctx:      context.Background(),
		runner:   cronRunner,
		execRepo: execRepo,
		jobRepo:  &testJobRepo{job: job},
		notifier: notify.NewSender(nil),
	}
	// NewScheduler registers HandleTurn as a runner turn listener; the mock
	// replays that dispatch so the rerun's turn is observed like in production.
	cronRunner.onTurn = s.HandleTurn

	first := s.executeJobWithTrigger(job, agentsv1.CronExecutionTriggerType_CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE)
	if first.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
		t.Fatalf("first run status = %s, want waiting_input", first.GetStatus())
	}

	second := s.executeJobWithTrigger(job, agentsv1.CronExecutionTriggerType_CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE)
	if second.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_SUCCESS {
		t.Fatalf("second run status = %s, want success", second.GetStatus())
	}

	if len(cronRunner.ctxInfos) != 2 {
		t.Fatalf("runner calls = %d, want 2", len(cronRunner.ctxInfos))
	}
	if got := cronRunner.ctxInfos[1].GetSessionId(); got == first.GetSessionId() {
		t.Fatalf("rescheduled run reused session %q where an execution is still waiting for input", got)
	}

	got := execRepo.execs[first.GetId()]
	if got.GetStatus() != agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_WAITING_INPUT {
		t.Fatalf("waiting execution status = %s, want still waiting_input: a scheduler tick must not answer for the human", got.GetStatus())
	}
	if got.GetOutput() != "Deploy v2 to production?" {
		t.Fatalf("waiting execution output = %q, want the node's question untouched", got.GetOutput())
	}
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

	if cronRunner.runCalls != 1 {
		t.Fatalf("runner calls = %d, want 1", cronRunner.runCalls)
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
	if cronRunner.runCalls != 1 {
		t.Fatalf("runner calls = %d, want 1", cronRunner.runCalls)
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
	if cronRunner.runCalls != 2 {
		t.Fatalf("runner calls = %d, want 2", cronRunner.runCalls)
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
	if cronRunner.runCalls != 1 {
		t.Fatalf("runner calls = %d, want 1", cronRunner.runCalls)
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

	// A cancellation (e.g. session deleted under a waiting execution, #132)
	// counts as failure for the notify policy.
	cancelled := &agentsv1.CronExecution{Status: agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_CANCELLED}
	if !shouldDeliver(&agentsv1.CronJob{NotifyOn: agentsv1.CronNotifyOn_CRON_NOTIFY_ON_FAILURE}, cancelled) {
		t.Fatal("failure-only policy should deliver cancelled execution")
	}
	if shouldDeliver(&agentsv1.CronJob{NotifyOn: agentsv1.CronNotifyOn_CRON_NOTIFY_ON_SUCCESS}, cancelled) {
		t.Fatal("success-only policy should not deliver cancelled execution")
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
