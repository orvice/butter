package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/notify"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	runtimeautomation "go.orx.me/apps/butter/internal/runtime/automation"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const (
	automationWorkspace      = "automation"
	automationAgentID        = "auto-agent"
	automationNotifyGroup    = "alerts"
	automationTelegramDestID = "auto-destination"
	automationTelegramChan   = "auto-channel"
	automationSeededManual   = "seeded-manual"
	automationCondFail       = "cond-fail"
	automationWebhookFail    = "webhook-fail"
)

type contractHTTPClient struct {
	mu       sync.Mutex
	statuses []int
}

func (c *contractHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	status := http.StatusOK
	if len(c.statuses) > 0 {
		status = c.statuses[0]
		c.statuses = c.statuses[1:]
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("ok")),
		Header:     make(http.Header),
	}, nil
}

type contractNotifier struct {
	mu    sync.Mutex
	calls int
}

func (n *contractNotifier) Send(context.Context, string, *agentsv1.NotifyTarget, notify.Message) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls++
	return nil
}

func registerAutomationFixture(
	ctx context.Context,
	r *gin.Engine,
	opts []connect.HandlerOption,
	agentRepo *configmemory.Store,
	runnerSvc *runner.Service,
	telegramRepo telegramrepo.Repository,
) error {
	if _, err := agentRepo.CreateAgent(ctx, automationWorkspace, &agentsv1.Agent{
		AgentId: automationAgentID, Name: "auto-agent", WorkspaceId: automationWorkspace,
		Config: &agentsv1.AgentConfig{Model: chatModel},
	}); err != nil {
		return err
	}

	if _, err := agentRepo.CreateNotifyGroup(ctx, automationWorkspace, &agentsv1.NotifyGroup{
		Name: automationNotifyGroup, WorkspaceId: automationWorkspace, Enabled: true,
		Targets: []*agentsv1.NotifyTarget{
			{
				Name: "lark", Enabled: true,
				Type: agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
				Lark: &agentsv1.LarkNotifyTarget{WebhookUrl: "https://example.test/lark"},
			},
		},
	}); err != nil {
		return err
	}

	if _, err := telegramRepo.CreateChannel(ctx, automationWorkspace,
		&agentsv1.TelegramChannel{
			Id: automationTelegramChan, Key: "auto", Name: "auto", BotId: "999002",
		},
		telegramrepo.ChannelCredentials{BotToken: telegramrepo.Credential{Ciphertext: "unused", KeyID: "contract"}}); err != nil {
		return err
	}
	if _, err := telegramRepo.CreateDestination(ctx, automationWorkspace, &agentsv1.TelegramDestination{
		Id: automationTelegramDestID, Key: "auto-topic", ChannelId: automationTelegramChan, ChatId: "-10088",
		OutboundEnabled: true,
		Config:          &agentsv1.TelegramDestinationConfig{AgentId: automationAgentID},
	}); err != nil {
		return err
	}

	defRepo := runtimeautomation.NewMemoryDefinitionRepo()
	runRepo := runtimeautomation.NewMemoryRunRepo()
	stepRepo := runtimeautomation.NewMemoryStepRunRepo()

	httpClient := &contractHTTPClient{statuses: []int{http.StatusOK, http.StatusInternalServerError}}
	notifier := &contractNotifier{}

	engine := runtimeautomation.NewEngine(defRepo, runRepo, stepRepo, runtimeautomation.EngineOptions{
		Runner:          runnerSvc,
		NotifyGroupRepo: agentRepo,
		Notifier:        notifier,
		HTTPClient:      httpClient,
		Context:         ctx,
	})

	sched, err := runtimeautomation.NewScheduler(ctx, defRepo, engine)
	if err != nil {
		return err
	}
	sched.Start()

	autoSvc := application.NewAutomationServiceServer()
	autoSvc.SetRepos(defRepo, runRepo, stepRepo)
	autoSvc.SetEngine(engine)
	autoSvc.SetScheduler(sched)
	autoSvc.SetAgentValidator(runnerSvc)

	manual := &agentsv1.Automation{
		Name: automationSeededManual, Enabled: true, WorkspaceId: automationWorkspace,
		Trigger: &agentsv1.AutomationTrigger{
			Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL,
		},
		Steps: []*agentsv1.AutomationStep{
			{
				Name: "hook", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK,
				CallWebhook: &agentsv1.AutomationCallWebhookStep{
					Url: "https://contract.test/hook", Method: http.MethodPost, PayloadJson: `{"ok":true}`,
				},
			},
			{
				Name: "notify", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP,
				SendNotifyGroup: &agentsv1.AutomationSendNotifyGroupStep{
					NotifyGroupName: automationNotifyGroup, Title: "done", Message: "ok",
				},
			},
		},
	}
	if err := defRepo.Create(ctx, manual); err != nil {
		return err
	}

	condFail := &agentsv1.Automation{
		Name: automationCondFail, Enabled: true, WorkspaceId: automationWorkspace,
		Trigger: &agentsv1.AutomationTrigger{
			Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL,
		},
		Conditions: []*agentsv1.AutomationCondition{
			{
				Selector: "payload.kind", Operator: agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_EQUALS,
				Value: "incident",
			},
		},
		Steps: []*agentsv1.AutomationStep{
			{
				Name: "hook", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK,
				CallWebhook: &agentsv1.AutomationCallWebhookStep{Url: "https://contract.test/skip"},
			},
		},
	}
	if err := defRepo.Create(ctx, condFail); err != nil {
		return err
	}

	webhookFail := &agentsv1.Automation{
		Name: automationWebhookFail, Enabled: true, WorkspaceId: automationWorkspace,
		Trigger: &agentsv1.AutomationTrigger{
			Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL,
		},
		Steps: []*agentsv1.AutomationStep{
			{
				Name: "hook", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK,
				CallWebhook: &agentsv1.AutomationCallWebhookStep{Url: "https://contract.test/fail"},
			},
		},
	}
	if err := defRepo.Create(ctx, webhookFail); err != nil {
		return err
	}

	autoPath, autoHandler := agentsv1connect.NewAutomationServiceHandler(autoSvc, opts...)
	r.Any("/api"+autoPath+"*path", gin.WrapH(http.StripPrefix("/api", autoHandler)))

	return nil
}
