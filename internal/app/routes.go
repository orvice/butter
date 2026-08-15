package app

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/mcpoauth"
	workspacemcp "go.orx.me/apps/butter/internal/mcpserver"
	"go.orx.me/apps/butter/internal/repo"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/asyncrun"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	telegramruntime "go.orx.me/apps/butter/internal/runtime/telegram"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/service"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramqueue"
	"go.orx.me/apps/butter/internal/telegramsend"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// Handlers holds all HTTP and ConnectRPC handlers that need post-bootstrap wiring.
type Handlers struct {
	a2aHandler             *httpHandler.A2AHandler
	openAIHandler          *httpHandler.OpenAIHandler
	aguiHandler            *httpHandler.AGUIHandler
	forumSvcServer         *application.ForumServiceServer
	agentSvcServer         *application.AgentServiceServer
	agentFileSvcServer     *application.AgentFileServiceServer
	skillSvcServer         *application.SkillServiceServer
	mcpSvcServer           *application.MCPServerServiceServer
	modelProviderSvcServer *application.ModelProviderServiceServer
	notifyGroupSvcServer   *application.NotifyGroupServiceServer
	remoteSvcServer        *application.RemoteAgentServiceServer
	automationSvcServer    *application.AutomationServiceServer
	sessionSvcServer       *application.SessionServiceServer
	cronSvcServer          *application.CronJobServiceServer
	channelSvcServer       *application.ChannelServiceServer
	dashboardSvcServer     *application.DashboardServiceServer
	daemonSvcServer        *application.DaemonServiceServer
	daemonConnectorHandler *daemon.GRPCHandler
	apiTokenSvcServer      *application.APITokenServiceServer
	authSvcServer          *application.AuthServiceServer
	workspaceSvcServer     *application.WorkspaceServiceServer
	gitHostSvcServer       *application.GitHostServiceServer
	repoBindingSvcServer   *application.RepoBindingServiceServer
	tgChannelSvcServer     *application.TelegramChannelServiceServer
	tgDestinationSvcServer *application.TelegramDestinationServiceServer
	tgAdminSvcServer       *application.TelegramAdminServiceServer
	tgProcessingSvcServer  *application.TelegramProcessingServiceServer
	tgReceiver             atomic.Value // *telegram.Receiver
	tgReconciler           *telegramruntime.Reconciler
	tgWorker               *telegramruntime.Worker
	tgPoller               *telegramruntime.Poller
	workspaceMCPSvc        *workspacemcp.Service
	authRepo               atomic.Value // auth.Repository
	apiTokenRepo           atomic.Value // apitoken.Repository
	forumRepo              atomic.Value // forum.Repository
	workspaceRepo          atomic.Value // workspace.Repository
	configStore            *ConfigStore
	configRuntime          *ConfigRuntime
	agentRepo              configrepo.AgentRepository
	mcpServerRepo          configrepo.MCPServerRepository
	modelProviderRepo      configrepo.ModelProviderRepository
	notifyGroupRepo        configrepo.NotifyGroupRepository
	remoteAgentRepo        configrepo.RemoteAgentRepository
	channelRepo            configrepo.ChannelRepository
	cfg                    *config.AppConfig
	reconciler             *Reconciler
	asyncCoordinator       *asyncrun.Coordinator
}

// Authenticate implements httpHandler.TelegramReceiver. The public callback
// route is registered before bootstrap wires the receiver, so the lookup is
// atomic and a request that arrives too early is told to retry rather than
// panicking.
func (h *Handlers) Authenticate(ctx *gin.Context, channelID, secret string) (telegramruntime.AuthenticatedChannel, error) {
	receiver := h.telegramReceiver()
	if receiver == nil {
		return telegramruntime.AuthenticatedChannel{}, errTelegramReceiveNotReady
	}
	return receiver.Authenticate(ctx.Request.Context(), channelID, secret)
}

// Deliver implements httpHandler.TelegramReceiver.
func (h *Handlers) Deliver(ctx *gin.Context, channel telegramruntime.AuthenticatedChannel, raw []byte) (telegramruntime.Decision, error) {
	receiver := h.telegramReceiver()
	if receiver == nil {
		return telegramruntime.DecisionIgnored, errTelegramReceiveNotReady
	}
	return receiver.Deliver(ctx.Request.Context(), channel, raw)
}

var errTelegramReceiveNotReady = errors.New("telegram receive is not ready")

// telegramFileClientFactory builds a per-request Telegram file client from
// the Channel's current credential. Like the sender, it caches nothing: a
// rotated token must take effect on the next download.
func telegramFileClientFactory(repo telegramrepo.Repository, keyring *secretbox.Keyring) func(context.Context, string, string) (telegramapi.FileClient, error) {
	return func(ctx context.Context, workspaceID, channelID string) (telegramapi.FileClient, error) {
		cred, err := repo.GetChannelCredential(ctx, workspaceID, channelID)
		if err != nil {
			return nil, err
		}
		token, err := keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
		if err != nil {
			return nil, err
		}
		return telegramapi.New(string(token)), nil
	}
}

const (
	// telegramLeaseTTL bounds how long a crashed leader blocks the fleet.
	telegramLeaseTTL = 30 * time.Second
	// telegramReconcileInterval is how often the leader compares desired and
	// observed Webhook registration. Reconciliation is a read unless they
	// differ, so a short interval costs little.
	telegramReconcileInterval = 30 * time.Second
)

func (h *Handlers) telegramReceiver() *telegramruntime.Receiver {
	if h == nil {
		return nil
	}
	v := h.tgReceiver.Load()
	if v == nil {
		return nil
	}
	receiver, _ := v.(*telegramruntime.Receiver)
	return receiver
}

// StopTelegramRuntime halts the reconciler and worker for a graceful exit.
func (h *Handlers) StopTelegramRuntime() {
	if h == nil {
		return
	}
	if h.tgPoller != nil {
		h.tgPoller.Stop()
	}
	if h.tgWorker != nil {
		h.tgWorker.Stop()
	}
	if h.tgReconciler != nil {
		h.tgReconciler.Stop()
	}
}

// ShutdownAsync stops process-owned async dashboard work for a graceful
// process exit. In-flight runs persist an honest FAILED terminal state; the
// call blocks until those writes complete or ctx expires.
func (h *Handlers) ShutdownAsync(ctx context.Context) error {
	if h == nil || h.asyncCoordinator == nil {
		return nil
	}
	return h.asyncCoordinator.Shutdown(ctx)
}

// apiTokenRepoFromHolder returns the currently wired apitoken repository, if any.
func (h *Handlers) apiTokenRepoFromHolder() apitoken.Repository {
	if h == nil {
		return nil
	}
	v := h.apiTokenRepo.Load()
	if v == nil {
		return nil
	}
	repo, _ := v.(apitoken.Repository)
	return repo
}

func (h *Handlers) authRepoFromHolder() auth.Repository {
	if h == nil {
		return nil
	}
	v := h.authRepo.Load()
	if v == nil {
		return nil
	}
	repo, _ := v.(auth.Repository)
	return repo
}

func (h *Handlers) workspaceRepoFromHolder() workspace.Repository {
	if h == nil {
		return nil
	}
	v := h.workspaceRepo.Load()
	if v == nil {
		return nil
	}
	repo, _ := v.(workspace.Repository)
	return repo
}

// Wire connects the bootstrap result to the handlers.
func (h *Handlers) Wire(result *BootstrapResult) {
	if result == nil {
		return
	}
	if result.RunnerSvc != nil {
		h.a2aHandler.SetRunnerService(result.RunnerSvc)
		h.openAIHandler.SetRunnerService(result.RunnerSvc)
		h.aguiHandler.SetRunnerService(result.RunnerSvc)
		h.sessionSvcServer.SetRunnerService(result.RunnerSvc)
		h.agentSvcServer.SetRunnerService(result.RunnerSvc)
		if h.forumSvcServer != nil {
			h.forumSvcServer.SetRunnerService(result.RunnerSvc)
		}
	}
	if result.InvocationRepo != nil {
		h.agentSvcServer.SetInvocationRepo(result.InvocationRepo)
		if h.dashboardSvcServer != nil {
			h.dashboardSvcServer.SetInvocationRepo(result.InvocationRepo)
		}
	}
	if result.InputPartRepo != nil {
		h.agentSvcServer.SetInputPartRepo(result.InputPartRepo)
	}
	if result.AsyncCoordinator != nil {
		h.asyncCoordinator = result.AsyncCoordinator
		h.agentSvcServer.SetAsyncCoordinator(result.AsyncCoordinator)
		h.sessionSvcServer.SetAsyncCoordinator(result.AsyncCoordinator)
		// Wire best-effort title generation after the first successful async turn.
		if h.sessionSvcServer != nil {
			sessServer := h.sessionSvcServer
			result.AsyncCoordinator.SetTurnCompleteCallback(sessServer.AsyncTurnComplete)
		}
	}
	if result.InputPartRepo != nil {
		h.sessionSvcServer.SetInputPartRepo(result.InputPartRepo)
	}
	// Wire the session exclusion checker so the submit path rejects new
	// invocations on sessions being deleted (issue #252).
	h.agentSvcServer.SetSessionExcluder(h.sessionSvcServer)
	if result.SessionSvc != nil {
		h.agentSvcServer.SetSessionService(result.SessionSvc)
	}
	if result.MCPOAuthSvc != nil {
		h.mcpSvcServer.SetOAuthService(result.MCPOAuthSvc)
	}
	if result.AgentFileRepo != nil && h.agentFileSvcServer != nil {
		h.agentFileSvcServer.SetRepo(result.AgentFileRepo)
		h.agentFileSvcServer.SetMaxFileBytes(result.AgentFileMaxBytes)
	}
	if result.SkillRepo != nil && h.skillSvcServer != nil {
		h.skillSvcServer.SetRepo(result.SkillRepo)
		h.skillSvcServer.SetSkillMDMaxBytes(result.SkillMDMaxBytes)
		h.skillSvcServer.SetSkillResourceMaxCount(result.SkillResourceMaxCount)
	}
	if result.AgentFileRepo != nil && h.workspaceMCPSvc != nil {
		h.workspaceMCPSvc.SetAgentFileRepo(result.AgentFileRepo)
	}
	if result.MCPAuthResolver != nil {
		h.mcpSvcServer.SetMCPHTTPClientFactory(result.MCPAuthResolver)
	}
	if result.ForumRepo != nil {
		h.forumRepo.Store(result.ForumRepo)
		if h.forumSvcServer != nil {
			h.forumSvcServer.SetRepo(result.ForumRepo)
		}
	}
	if result.LangfuseHost != "" {
		h.sessionSvcServer.SetLangfuseHost(result.LangfuseHost)
	}
	if result.SessionSvc != nil {
		h.sessionSvcServer.SetSessionService(result.SessionSvc)
	}
	if result.SessionTitleStore != nil {
		h.sessionSvcServer.SetTitleStore(result.SessionTitleStore)
	}
	if result.SessionWSStore != nil {
		h.sessionSvcServer.SetWorkspaceSessionStore(result.SessionWSStore)
	}
	if result.SessionReadStore != nil {
		h.sessionSvcServer.SetReadStore(result.SessionReadStore)
	}
	if result.InvocationRepo != nil {
		h.sessionSvcServer.SetInvocationRepo(result.InvocationRepo)
	}
	if result.RunnerSvc != nil {
		h.sessionSvcServer.SetTitleModelResolver(result.RunnerSvc)
	}
	if h.modelProviderRepo != nil {
		h.sessionSvcServer.SetTitleProviderLister(h.modelProviderRepo)
	}
	if result.ChatTitleModel != "" {
		h.sessionSvcServer.SetChatTitleModel(result.ChatTitleModel)
	}
	if result.CronScheduler != nil {
		h.cronSvcServer.SetScheduler(result.CronScheduler)
		// Deleting a paused session abandons its workflow (ADR 0002); the
		// scheduler cancels the session's WAITING_INPUT executions (#132).
		h.sessionSvcServer.AddSessionDeleteListener(result.CronScheduler.HandleSessionDeleted)
	}
	if result.CronRepo != nil {
		h.cronSvcServer.SetExecutionRepo(result.CronRepo)
	}
	if h.automationSvcServer != nil {
		if result.AutomationDefRepo != nil || result.AutomationRunRepo != nil || result.AutomationStepRepo != nil {
			h.automationSvcServer.SetRepos(result.AutomationDefRepo, result.AutomationRunRepo, result.AutomationStepRepo)
		}
		if result.AutomationEngine != nil {
			h.automationSvcServer.SetEngine(result.AutomationEngine)
			// Deleting a paused session abandons its workflow (ADR 0002); the
			// engine cancels the session's WAITING_INPUT automation run (#176).
			h.sessionSvcServer.AddSessionDeleteListener(result.AutomationEngine.HandleSessionDeleted)
		}
		if result.RunnerSvc != nil {
			h.automationSvcServer.SetAgentValidator(result.RunnerSvc)
		}
		if result.AutomationScheduler != nil {
			h.automationSvcServer.SetScheduler(result.AutomationScheduler)
		}
	}
	if h.configRuntime != nil {
		if result.RunnerSvc != nil {
			h.configRuntime.SetRunnerService(result.RunnerSvc)
		}
		if result.ChannelMgr != nil {
			h.configRuntime.SetChannelManager(result.ChannelMgr)
		}
		h.agentSvcServer.SetRuntime(h.configRuntime)
		h.mcpSvcServer.SetRuntime(h.configRuntime)
		h.modelProviderSvcServer.SetRuntime(h.configRuntime)
		h.remoteSvcServer.SetRuntime(h.configRuntime)
	}
	if result.APITokenRepo != nil {
		h.apiTokenRepo.Store(result.APITokenRepo)
		if h.apiTokenSvcServer != nil {
			h.apiTokenSvcServer.SetRepo(result.APITokenRepo)
		}
		if h.daemonSvcServer != nil {
			h.daemonSvcServer.SetAPITokenRepo(result.APITokenRepo)
		}
		if h.daemonConnectorHandler != nil {
			h.daemonConnectorHandler.SetAPITokenRepo(result.APITokenRepo)
		}
	}
	if result.AuthRepo != nil {
		h.authRepo.Store(result.AuthRepo)
		if h.authSvcServer != nil {
			h.authSvcServer.SetRepo(result.AuthRepo)
		}
	}
	if h.authSvcServer != nil {
		if result.OAuthStateRepo != nil {
			h.authSvcServer.SetOAuthStateRepo(result.OAuthStateRepo)
		}
		if result.OAuthProviders != nil {
			h.authSvcServer.SetProviderRegistry(result.OAuthProviders)
		}
	}
	if result.WorkspaceRepo != nil {
		h.workspaceRepo.Store(result.WorkspaceRepo)
		if h.workspaceSvcServer != nil {
			h.workspaceSvcServer.SetRepo(result.WorkspaceRepo)
		}
		if h.workspaceMCPSvc != nil {
			h.workspaceMCPSvc.SetWorkspaceRepo(result.WorkspaceRepo)
		}
		if h.authSvcServer != nil {
			h.authSvcServer.SetWorkspaceRepo(result.WorkspaceRepo)
		}
		h.agentSvcServer.SetWorkspaceRepo(result.WorkspaceRepo)
		if h.repoBindingSvcServer != nil {
			h.repoBindingSvcServer.SetWorkspaceRepo(result.WorkspaceRepo)
		}
	}
	// Consumer-record sources for the read-only VerifyAgentIDCutover RPC
	// (issue #241).
	cutoverSources := application.AgentCutoverSources{
		Agents:      h.configStore,
		Channels:    h.configStore,
		CronJobs:    result.CronJobRepo,
		Automations: result.AutomationDefRepo,
		Forum:       result.ForumRepo,
		Workspaces:  result.WorkspaceRepo,
	}
	if result.RunnerSvc != nil {
		cutoverSources.ReservedName = result.RunnerSvc.IsReservedAgentName
	}
	h.agentSvcServer.SetCutoverSources(cutoverSources)
	if h.gitHostSvcServer != nil && result.GitHostRepo != nil {
		h.gitHostSvcServer.SetRepo(result.GitHostRepo)
	}
	// Telegram Channels/Destinations (issue #264). The keyring is shared by
	// both services so a Bot Token encrypted by one is readable by the other.
	if result.TelegramRepo != nil {
		keyring := secretbox.NewKeyring(result.CryptoKeyRepo)
		if h.tgChannelSvcServer != nil {
			h.tgChannelSvcServer.SetRepo(result.TelegramRepo)
			h.tgChannelSvcServer.SetKeyring(keyring)
		}
		// One sender instance backs every outbound Telegram path: Dashboard
		// test messages, Notify Groups, and Cron delivery.
		sender := telegramsend.New(result.TelegramRepo, keyring, nil)
		if h.tgDestinationSvcServer != nil {
			h.tgDestinationSvcServer.SetRepo(result.TelegramRepo)
			h.tgDestinationSvcServer.SetSender(sender)
			h.tgDestinationSvcServer.SetReferenceRepos(h.notifyGroupRepo, result.CronJobRepo)
			if store := telegramruntime.NewRedisPreferenceStore(result.Redis); store != nil {
				h.tgDestinationSvcServer.SetPreferenceCleaner(store)
			}
		}
		if h.notifyGroupSvcServer != nil {
			h.notifyGroupSvcServer.SetTelegramRepo(result.TelegramRepo)
		}
		// Telegram Destinations hold strong references to Agents and Models;
		// the guard turns "the topic silently stopped working" into a
		// deliberate decision at delete time (#264).
		guard := application.NewTelegramReferenceGuard(result.TelegramRepo)
		h.agentSvcServer.SetTelegramGuard(guard)
		if h.modelProviderSvcServer != nil {
			h.modelProviderSvcServer.SetTelegramGuard(guard)
		}
		if h.tgProcessingSvcServer != nil && result.TelegramProcessingRepo != nil {
			h.tgProcessingSvcServer.SetRepo(result.TelegramProcessingRepo)
			h.tgProcessingSvcServer.SetSender(sender)
		}
		if h.tgAdminSvcServer != nil && result.TelegramSettingRepo != nil {
			h.tgAdminSvcServer.SetRepo(result.TelegramSettingRepo)
		}
		if h.tgChannelSvcServer != nil && result.TelegramSettingRepo != nil {
			h.tgChannelSvcServer.SetSettingsRepo(result.TelegramSettingRepo)
		}

		// The receive runtime only exists when Redis is available: Webhook
		// mode treats Redis as durable queue infrastructure, not a cache, so
		// standing it up without one would accept updates nothing can hold.
		queue := telegramqueue.New(result.Redis)
		if queue.Available() {
			h.tgChannelSvcServer.SetQueueProbe(queue)
			router := telegramruntime.NewRouter(result.TelegramRepo, queue)
			h.tgReceiver.Store(telegramruntime.NewReceiver(result.TelegramRepo, keyring, router))

			instanceID := uuid.NewString()
			lease := telegramqueue.NewLease(result.Redis,
				telegramqueue.WebhookReconcilerLeaseKey, instanceID, telegramLeaseTTL)
			h.tgReconciler = telegramruntime.NewReconciler(
				result.TelegramRepo, result.TelegramSettingRepo, keyring, nil, lease,
				telegramReconcileInterval)
			h.tgReconciler.Start(context.Background())
			h.tgChannelSvcServer.SetWebhookStatusSource(h.tgReconciler)

			// `/where` is answered by the transport-level handler; everything
			// else falls through to the Destination-scoped orchestrator.
			var interactions telegramruntime.EventHandler
			if result.RunnerSvc != nil {
				orchestrator := telegramruntime.NewOrchestrator(result.TelegramRepo, sender, result.RunnerSvc)
				if result.SessionSvc != nil {
					orchestrator.SetSessionClearer(result.SessionSvc)
				}
				// Selections live in Redis rather than process memory: any Pod
				// may handle the next message for the same destination, and a
				// choice must outlive a restart.
				orchestrator.SetPreferenceStore(telegramruntime.NewRedisPreferenceStore(result.Redis))
				// Media is downloaded on the worker immediately before
				// invocation, so Redis never holds binary data.
				orchestrator.SetFileClientFactory(telegramFileClientFactory(result.TelegramRepo, keyring))
				// The durable state machine and the session lease are what
				// make a crashed worker's work reclaimable without repeating
				// agent side effects.
				if result.TelegramProcessingRepo != nil {
					orchestrator.SetProcessingRepo(result.TelegramProcessingRepo)
				}
				orchestrator.SetSessionGuard(telegramruntime.NewRedisSessionGuard(result.Redis, instanceID))
				interactions = orchestrator
			}
			h.tgWorker = telegramruntime.NewWorker(queue,
				telegramruntime.NewWhereHandler(sender, interactions), instanceID)
			if err := h.tgWorker.Start(context.Background()); err != nil {
				h.tgWorker = nil
			}

			// Long Polling is an alternative transport, not an alternative
			// pipeline: it feeds the same router, queue, and workers.
			h.tgPoller = telegramruntime.NewPoller(
				result.TelegramRepo, keyring, router,
				telegramruntime.NewRedisOffsetStore(result.Redis), nil,
				telegramruntime.RedisPollingLeaseFactory(result.Redis, instanceID),
				instanceID)
			h.tgPoller.Start(context.Background())
			h.tgChannelSvcServer.SetPollingStatusSource(h.tgPoller)
		}
		if h.cronSvcServer != nil {
			h.cronSvcServer.SetTelegramRepo(result.TelegramRepo)
		}
		if result.CronScheduler != nil {
			result.CronScheduler.SetTelegramDelivery(sender)
		}
		if result.AutomationEngine != nil {
			result.AutomationEngine.SetTelegramDelivery(sender)
		}
	}
	if result.WorkspaceRepo != nil {
		if h.tgChannelSvcServer != nil {
			h.tgChannelSvcServer.SetWorkspaceRepo(result.WorkspaceRepo)
		}
		if h.tgDestinationSvcServer != nil {
			h.tgDestinationSvcServer.SetWorkspaceRepo(result.WorkspaceRepo)
		}
		if h.tgProcessingSvcServer != nil {
			h.tgProcessingSvcServer.SetWorkspaceRepo(result.WorkspaceRepo)
		}
	}
	if h.repoBindingSvcServer != nil && result.RepoBindingRepo != nil && result.GitHostRepo != nil {
		h.repoBindingSvcServer.SetRepos(result.RepoBindingRepo, result.GitHostRepo)
	}
	if h.repoBindingSvcServer != nil && result.RepoCacheRepo != nil {
		h.repoBindingSvcServer.SetCacheRepo(result.RepoCacheRepo)
	}
	if h.repoBindingSvcServer != nil && result.AgentContentRepo != nil {
		h.repoBindingSvcServer.SetContentRepo(result.AgentContentRepo)
		h.repoBindingSvcServer.SetConfigRuntime(h.configRuntime)
	}
	// Wire the Agent lifecycle Saga (issue #218): the durable operation store
	// and the Git content seam (satisfied by the repo-binding service). Only
	// when both are present does CreateAgent/DeleteAgent take the Saga path.
	if result.AgentOperationRepo != nil {
		h.agentSvcServer.SetOperationRepo(result.AgentOperationRepo)
	}
	if h.repoBindingSvcServer != nil {
		h.agentSvcServer.SetContentCoordinator(h.repoBindingSvcServer)
	}
	if h.configRuntime != nil && result.RepoBindingRepo != nil && result.AgentContentRepo != nil {
		h.configRuntime.SetAgentContentRepos(result.RepoBindingRepo, result.AgentContentRepo)
	}
	if h.repoBindingSvcServer != nil && result.RepoBindingRepo != nil {
		reconciler := NewReconciler(result.RepoBindingRepo, h.repoBindingSvcServer, h.cfg.Git.EffectiveReconcileInterval())
		h.reconciler = reconciler
		reconciler.Start(context.Background())
	}
	if h.dashboardSvcServer != nil {
		if result.MongoDB != nil {
			h.dashboardSvcServer.SetMongo(result.MongoDB)
		}
		if result.Redis != nil {
			h.dashboardSvcServer.SetRedis(result.Redis)
		}
		if result.CronJobRepo != nil {
			h.dashboardSvcServer.SetCronJobRepo(result.CronJobRepo)
		}
		if result.CronRepo != nil {
			h.dashboardSvcServer.SetCronExecutionRepo(result.CronRepo)
		}
		if result.SessionCounter != nil {
			h.dashboardSvcServer.SetSessionCounter(result.SessionCounter)
		}
		if result.RunnerSvc != nil {
			runner := result.RunnerSvc
			h.dashboardSvcServer.SetRunnerReady(func() bool { return runner != nil })
		}
	}
}

// StopReconciler stops the background reconciler if running.
func (h *Handlers) StopReconciler() {
	if h.reconciler != nil {
		h.reconciler.Stop()
	}
}

// AgentRepo returns the agent repository.
func (h *Handlers) AgentRepo() configrepo.AgentRepository {
	return h.agentRepo
}

// ChannelRepo returns the channel repository.
func (h *Handlers) ChannelRepo() configrepo.ChannelRepository {
	return h.channelRepo
}

// NotifyGroupRepo returns the notify group repository.
func (h *Handlers) NotifyGroupRepo() configrepo.NotifyGroupRepository {
	return h.notifyGroupRepo
}

// ConfigStore returns the shared workspace-scoped configuration repository.
func (h *Handlers) ConfigStore() *ConfigStore {
	return h.configStore
}

// SeedConfig initializes and seeds the config repositories from AppConfig.
func (h *Handlers) SeedConfig(ctx context.Context, cfg *config.AppConfig) error {
	if h.configStore == nil {
		return nil
	}
	return h.configStore.InitFromConfig(ctx, cfg)
}

// SetupRoutes creates all handlers and returns a Gin router function plus
// the Handlers struct for post-bootstrap wiring.
func SetupRoutes(cfg *config.AppConfig, daemonRegistry *daemon.Registry) (func(r *gin.Engine), *Handlers) {
	configStore := NewConfigStore()
	configRuntime := NewConfigRuntime(configStore, cfg)

	healthRepo := repo.NewHealthRepository()
	healthService := service.NewHealthService(healthRepo, cfg)
	healthHandler := httpHandler.NewHealthHandler(healthService)
	statusService := service.NewStatusService(cfg, configStore)
	statusHandler := httpHandler.NewStatusHandler(statusService)
	a2aHandler := httpHandler.NewA2AHandler(cfg)
	openAIHandler := httpHandler.NewOpenAIHandler(configStore)
	aguiHandler := httpHandler.NewAGUIHandler(configStore)
	// Lazy provider: SetupRoutes runs before core.New loads YAML into cfg,
	// so we read cfg.Static on every request instead of snapshotting now.
	uploadSvc := service.NewUploadServiceLazy(func() config.StaticConfig { return cfg.Static })
	uploadHandler := httpHandler.NewUploadHandler(uploadSvc)

	agentSvcServer := application.NewAgentServiceServer(configStore)
	agentFileSvcServer := application.NewAgentFileServiceServer(nil)
	skillSvcServer := application.NewSkillServiceServer(nil)
	mcpSvcServer := application.NewMCPServerServiceServer(configStore)
	modelProviderSvcServer := application.NewModelProviderServiceServer(configStore)
	notifyGroupSvcServer := application.NewNotifyGroupServiceServer(configStore)
	remoteSvcServer := application.NewRemoteAgentServiceServer(configStore)
	remoteSvcServer.SetDaemonRegistry(daemonRegistry)
	automationSvcServer := application.NewAutomationServiceServer()
	// Every Connect handler shares the same option set so the wire format
	// matches the pre-migration Twirp behavior (snake_case JSON). Without
	// this, dashboard callers that still read response fields like
	// connected_daemons / base_url / space_id without camelCase fallbacks
	// silently see undefined.
	connectOpts := connectx.HandlerOptions()
	forumSvcServer := application.NewForumServiceServer(nil)
	forumConnectPath, forumConnectHandler := agentsv1connect.NewForumServiceHandler(forumSvcServer, connectOpts...)
	agentConnectPath, agentConnectHandler := agentsv1connect.NewAgentServiceHandler(agentSvcServer, connectOpts...)
	agentFileConnectPath, agentFileConnectHandler := agentsv1connect.NewAgentFileServiceHandler(agentFileSvcServer, connectOpts...)
	skillConnectPath, skillConnectHandler := agentsv1connect.NewSkillServiceHandler(skillSvcServer, connectOpts...)
	mcpConnectPath, mcpConnectHandler := agentsv1connect.NewMCPServerServiceHandler(mcpSvcServer, connectOpts...)
	modelProviderConnectPath, modelProviderConnectHandler := agentsv1connect.NewModelProviderServiceHandler(modelProviderSvcServer, connectOpts...)
	notifyGroupConnectPath, notifyGroupConnectHandler := agentsv1connect.NewNotifyGroupServiceHandler(notifyGroupSvcServer, connectOpts...)
	remoteConnectPath, remoteConnectHandler := agentsv1connect.NewRemoteAgentServiceHandler(remoteSvcServer, connectOpts...)
	automationConnectPath, automationConnectHandler := agentsv1connect.NewAutomationServiceHandler(automationSvcServer, connectOpts...)
	channelSvcServer := application.NewChannelServiceServer(configStore)
	channelSvcServer.SetAgentRepo(configStore)
	channelConnectPath, channelConnectHandler := agentsv1connect.NewChannelServiceHandler(channelSvcServer, connectOpts...)
	sessionSvcServer := application.NewSessionServiceServer()
	sessionConnectPath, sessionConnectHandler := agentsv1connect.NewSessionServiceHandler(sessionSvcServer, connectOpts...)
	cronSvcServer := application.NewCronJobServiceServer()
	cronConnectPath, cronConnectHandler := agentsv1connect.NewCronJobServiceHandler(cronSvcServer, connectOpts...)
	dashboardSvcServer := application.NewDashboardServiceServer(configStore, daemonRegistry)
	dashboardConnectPath, dashboardConnectHandler := agentsv1connect.NewDashboardServiceHandler(dashboardSvcServer, connectOpts...)
	daemonSvcServer := application.NewDaemonServiceServer(configStore, daemonRegistry)
	daemonConnectPath, daemonConnectHandler := agentsv1connect.NewDaemonServiceHandler(daemonSvcServer, connectOpts...)
	daemonConnectorHandler := daemon.NewGRPCHandler(daemonRegistry, nil, configStore)
	daemonConnectorConnectPath, daemonConnectorConnectHandler := agentsv1connect.NewDaemonConnectorServiceHandler(daemonConnectorHandler, connectOpts...)
	apiTokenSvcServer := application.NewAPITokenServiceServer(nil)
	apiTokenConnectPath, apiTokenConnectHandler := agentsv1connect.NewAPITokenServiceHandler(apiTokenSvcServer, connectOpts...)
	globalMCPSvcServer := application.NewGlobalMCPServerServiceServer(configStore, mcpSvcServer)
	globalMCPConnectPath, globalMCPConnectHandler := agentsv1connect.NewGlobalMCPServerServiceHandler(globalMCPSvcServer, connectOpts...)
	authSvcServer := application.NewAuthServiceServer(nil, cfg.Auth.EffectiveSessionTTL())
	authConnectPath, authConnectHandler := agentsv1connect.NewAuthServiceHandler(authSvcServer, connectOpts...)
	workspaceSvcServer := application.NewWorkspaceServiceServer(nil)
	workspaceConnectPath, workspaceConnectHandler := agentsv1connect.NewWorkspaceServiceHandler(workspaceSvcServer, connectOpts...)
	gitHostSvcServer := application.NewGitHostServiceServer(nil)
	gitHostConnectPath, gitHostConnectHandler := agentsv1connect.NewGitHostServiceHandler(gitHostSvcServer, connectOpts...)
	repoBindingSvcServer := application.NewRepoBindingServiceServer(nil, nil)
	repoBindingSvcServer.SetAgentRepo(configStore)
	// Lazy provider: SetupRoutes runs before core.New loads YAML into cfg.
	repoBindingSvcServer.SetEncryptionKeyProvider(func() string { return cfg.Git.EncryptionKey })
	repoBindingSvcServer.SetCacheLimitsProvider(func() (int64, int64) {
		return cfg.Git.EffectiveMaxFileBytes(), cfg.Git.EffectiveMaxWorkspaceCacheBytes()
	})
	repoBindingSvcServer.SetWebhookBaseURL(func() string { return cfg.Git.WebhookBaseURL })
	repoBindingConnectPath, repoBindingConnectHandler := agentsv1connect.NewWorkspaceRepoBindingServiceHandler(repoBindingSvcServer, connectOpts...)
	tgChannelSvcServer := application.NewTelegramChannelServiceServer(nil)
	tgChannelConnectPath, tgChannelConnectHandler := agentsv1connect.NewTelegramChannelServiceHandler(tgChannelSvcServer, connectOpts...)
	tgDestinationSvcServer := application.NewTelegramDestinationServiceServer(nil)
	tgDestinationSvcServer.SetConfigRepos(configStore, configStore)
	tgDestinationConnectPath, tgDestinationConnectHandler := agentsv1connect.NewTelegramDestinationServiceHandler(tgDestinationSvcServer, connectOpts...)
	tgAdminSvcServer := application.NewTelegramAdminServiceServer(nil)
	tgAdminConnectPath, tgAdminConnectHandler := agentsv1connect.NewTelegramAdminServiceHandler(tgAdminSvcServer, connectOpts...)
	tgProcessingSvcServer := application.NewTelegramProcessingServiceServer(nil)
	tgProcessingConnectPath, tgProcessingConnectHandler := agentsv1connect.NewTelegramProcessingServiceHandler(tgProcessingSvcServer, connectOpts...)
	workspaceMCPSvc := workspacemcp.NewService(configStore)

	handlers := &Handlers{
		a2aHandler:             a2aHandler,
		openAIHandler:          openAIHandler,
		aguiHandler:            aguiHandler,
		forumSvcServer:         forumSvcServer,
		agentSvcServer:         agentSvcServer,
		agentFileSvcServer:     agentFileSvcServer,
		skillSvcServer:         skillSvcServer,
		mcpSvcServer:           mcpSvcServer,
		modelProviderSvcServer: modelProviderSvcServer,
		notifyGroupSvcServer:   notifyGroupSvcServer,
		remoteSvcServer:        remoteSvcServer,
		automationSvcServer:    automationSvcServer,
		sessionSvcServer:       sessionSvcServer,
		cronSvcServer:          cronSvcServer,
		channelSvcServer:       channelSvcServer,
		dashboardSvcServer:     dashboardSvcServer,
		daemonSvcServer:        daemonSvcServer,
		daemonConnectorHandler: daemonConnectorHandler,
		apiTokenSvcServer:      apiTokenSvcServer,
		authSvcServer:          authSvcServer,
		workspaceSvcServer:     workspaceSvcServer,
		gitHostSvcServer:       gitHostSvcServer,
		repoBindingSvcServer:   repoBindingSvcServer,
		tgChannelSvcServer:     tgChannelSvcServer,
		tgDestinationSvcServer: tgDestinationSvcServer,
		tgAdminSvcServer:       tgAdminSvcServer,
		tgProcessingSvcServer:  tgProcessingSvcServer,
		workspaceMCPSvc:        workspaceMCPSvc,
		configStore:            configStore,
		configRuntime:          configRuntime,
		agentRepo:              configStore,
		mcpServerRepo:          configStore,
		modelProviderRepo:      configStore,
		notifyGroupRepo:        configStore,
		remoteAgentRepo:        configStore,
		channelRepo:            configStore,
		cfg:                    cfg,
	}

	router := func(r *gin.Engine) {
		r.Use(httpHandler.AuthMiddleware(cfg, handlers.authRepoFromHolder, handlers.apiTokenRepoFromHolder, handlers.workspaceRepoFromHolder))
		healthHandler.Register(r)
		statusHandler.Register(r)
		a2aHandler.Register(r)
		openAIHandler.Register(r)
		aguiHandler.Register(r)
		uploadHandler.Register(r)
		httpHandler.RegisterWorkspaceMCP(r, workspaceMCPSvc.Handler(), handlers.workspaceRepoFromHolder)
		r.GET(mcpoauth.CallbackPath, func(c *gin.Context) {
			status := "error"
			serverID := ""
			target := oauthCallbackFallback(cfg)
			if oauthErr := strings.TrimSpace(c.Query("error")); oauthErr != "" {
				c.Redirect(http.StatusFound, appendOAuthCallbackParams(target, status, serverID))
				return
			}
			returnURL, oauthStatus, err := mcpSvcServer.CompleteMCPServerOAuthCallback(c.Request.Context(), c.Query("state"), c.Query("code"))
			if returnURL != "" {
				target = returnURL
			}
			if err == nil {
				status = "success"
				if oauthStatus != nil {
					serverID = oauthStatus.GetServerId()
				}
			}
			c.Redirect(http.StatusFound, appendOAuthCallbackParams(target, status, serverID))
		})

		// Connect handlers mount at /agents.v1.XxxService/ by default; we
		// strip the /api prefix before forwarding so the public URLs
		// (/api/agents.v1.XxxService/Method) stay stable across the
		// migration.
		r.Any("/api"+authConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", authConnectHandler)))
		r.Any("/api"+workspaceConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", workspaceConnectHandler)))
		r.Any("/api"+apiTokenConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", apiTokenConnectHandler)))
		r.Any("/api"+agentConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", agentConnectHandler)))
		r.Any("/api"+mcpConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", mcpConnectHandler)))
		r.Any("/api"+modelProviderConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", modelProviderConnectHandler)))
		r.Any("/api"+notifyGroupConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", notifyGroupConnectHandler)))
		r.Any("/api"+remoteConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", remoteConnectHandler)))
		r.Any("/api"+automationConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", automationConnectHandler)))
		r.Any("/api"+channelConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", channelConnectHandler)))
		r.Any("/api"+forumConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", forumConnectHandler)))
		r.Any("/api"+agentFileConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", agentFileConnectHandler)))
		r.Any("/api"+skillConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", skillConnectHandler)))
		r.Any("/api"+sessionConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", sessionConnectHandler)))
		r.Any("/api"+cronConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", cronConnectHandler)))
		r.Any("/api"+dashboardConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", dashboardConnectHandler)))
		r.Any("/api"+daemonConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", daemonConnectHandler)))
		r.Any("/api"+daemonConnectorConnectPath+"*path", gin.WrapH(enableFullDuplex(http.StripPrefix("/api", daemonConnectorConnectHandler))))
		r.Any("/api"+globalMCPConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", globalMCPConnectHandler)))
		r.Any("/api"+gitHostConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", gitHostConnectHandler)))
		r.Any("/api"+repoBindingConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", repoBindingConnectHandler)))
		r.Any("/api"+tgChannelConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", tgChannelConnectHandler)))
		r.Any("/api"+tgDestinationConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", tgDestinationConnectHandler)))
		r.Any("/api"+tgAdminConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", tgAdminConnectHandler)))
		r.Any("/api"+tgProcessingConnectPath+"*path", gin.WrapH(http.StripPrefix("/api", tgProcessingConnectHandler)))

		// The Telegram callback is public: it authenticates with the
		// per-Channel secret Telegram echoes, not with a Butter session, and
		// it must be reachable on every Pod behind the load balancer.
		httpHandler.NewTelegramWebhookHandler(handlers).Register(r)

		webhookHandler := httpHandler.NewWebhookHandler(repoBindingSvcServer)
		r.POST("/api/webhooks/repository/:workspace_id", webhookHandler.Handle)
	}

	return router, handlers
}

func enableFullDuplex(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = http.NewResponseController(w).EnableFullDuplex()
		handler.ServeHTTP(w, r)
	})
}

func oauthCallbackFallback(cfg *config.AppConfig) string {
	base := strings.TrimSpace(cfg.MCPOAuth.DashboardBaseURL)
	if base == "" {
		return "/mcp-servers"
	}
	return strings.TrimRight(base, "/") + "/mcp-servers"
}

func appendOAuthCallbackParams(raw, status, serverID string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	q.Set("mcp_oauth", status)
	if serverID != "" {
		q.Set("server_id", serverID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
