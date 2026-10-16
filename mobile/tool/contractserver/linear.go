package main

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/application"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	"go.orx.me/apps/butter/internal/repo/auth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	linearmemory "go.orx.me/apps/butter/internal/repo/linear/memory"
	linearprocessingmemory "go.orx.me/apps/butter/internal/repo/linearprocessing/memory"
	linearsettingmemory "go.orx.me/apps/butter/internal/repo/linearsetting/memory"
	linearstatememory "go.orx.me/apps/butter/internal/repo/linearstate/memory"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/secretbox"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	agentsv1connect "go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const (
	linearWorkspace       = "linear"
	linearMemberWorkspace = "linear-member"
	linearAgentID         = "linear-agent"
	linearClientSecret    = "lin-client-secret-value"
	linearWebhookSecret   = "lin-webhook-secret-value"
	linearFailedRecordID  = "linear-failed-record"
)

type linearQueueProbe struct{}

func (linearQueueProbe) Available() bool                       { return true }
func (linearQueueProbe) Ping(context.Context) error            { return nil }
func (linearQueueProbe) CheckReady(context.Context) error      { return nil }
func (linearQueueProbe) CheckLeaseReady(context.Context) error { return nil }

type contractLinearResender struct {
	mu    sync.Mutex
	calls int
}

func (r *contractLinearResender) Resend(_ context.Context, workspaceID, id string) (*agentsv1.LinearProcessingRecord, error) {
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()
	return &agentsv1.LinearProcessingRecord{
		Id: id, WorkspaceId: workspaceID,
		Status: agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED,
	}, nil
}

type linearFixture struct {
	fake     *contractLinearFake
	resender *contractLinearResender
}

func registerLinearFixture(
	ctx context.Context,
	r *gin.Engine,
	opts []connect.HandlerOption,
	agentRepo *configmemory.Store,
	wsRepo workspacerepo.Repository,
) (*linearFixture, error) {
	if _, err := agentRepo.CreateAgent(ctx, linearWorkspace, &agentsv1.Agent{
		AgentId: linearAgentID, Name: "linear-agent", WorkspaceId: linearWorkspace,
	}); err != nil {
		return nil, err
	}
	settings := linearsettingmemory.New()
	if _, err := settings.Put(ctx, &agentsv1.LinearSettings{PublicBaseUrl: "http://127.0.0.1"}); err != nil {
		return nil, err
	}
	repo := linearmemory.New()
	states := linearstatememory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	apps := application.NewLinearAppServiceServer(repo)
	apps.SetWorkspaceRepo(wsRepo)
	apps.SetAgentRepo(agentRepo)
	apps.SetKeyring(keyring)
	apps.SetSettingsRepo(settings)
	apps.SetInstallStateRepo(states)
	apps.SetQueueProbe(linearQueueProbe{})
	apps.SetDashboardBaseURL(func() string { return "http://127.0.0.1" })

	fake := newContractLinearFake()
	apps.SetLinearClient(fake.Client())

	admin := application.NewLinearAdminServiceServer(settings)
	processingRepo := linearprocessingmemory.New()
	resender := &contractLinearResender{}
	proc := application.NewLinearProcessingServiceServer(processingRepo)
	proc.SetWorkspaceRepo(wsRepo)
	proc.SetResender(resender)

	ownerCtx := wsctx.WithID(
		auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "mobile-test", Username: "mobile-test"}, nil),
		linearWorkspace,
	)
	app := mustCreateLinearApp(ownerCtx, apps, &agentsv1.LinearApp{
		DisplayName: "Seeded", ClientId: "contract-linear-client", AgentId: linearAgentID,
	}, linearClientSecret, linearWebhookSecret)
	_ = app

	now := time.Now()
	rec, _, err := processingRepo.Claim(ctx, &agentsv1.LinearProcessingRecord{
		Id: linearFailedRecordID, WorkspaceId: linearWorkspace, AppId: app.GetId(),
		DeliveryId: "delivery-failed", Status: agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED,
		Output: "persisted reply body",
	}, "", now, now)
	if err != nil {
		return nil, err
	}
	rec.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED
	if _, err := processingRepo.Update(ctx, rec); err != nil {
		return nil, err
	}

	httpHandler.NewLinearOAuthHandler(apps).Register(r)
	_, appHandler := agentsv1connect.NewLinearAppServiceHandler(apps, opts...)
	_, adminHandler := agentsv1connect.NewLinearAdminServiceHandler(admin, opts...)
	_, procHandler := agentsv1connect.NewLinearProcessingServiceHandler(proc, opts...)
	for _, procedure := range []string{
		agentsv1connect.LinearAppServiceListLinearAppsProcedure,
		agentsv1connect.LinearAppServiceGetLinearAppProcedure,
		agentsv1connect.LinearAppServiceCreateLinearAppProcedure,
		agentsv1connect.LinearAppServiceUpdateLinearAppProcedure,
		agentsv1connect.LinearAppServicePutLinearAppCredentialsProcedure,
		agentsv1connect.LinearAppServiceDeleteLinearAppProcedure,
		agentsv1connect.LinearAppServiceBeginLinearInstallProcedure,
		agentsv1connect.LinearAppServiceListLinearInstallationsProcedure,
		agentsv1connect.LinearAppServiceDeleteLinearInstallationProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", appHandler)))
	}
	r.POST("/api"+agentsv1connect.LinearAdminServiceGetLinearSettingsProcedure,
		gin.WrapH(http.StripPrefix("/api", adminHandler)))
	for _, procedure := range []string{
		agentsv1connect.LinearProcessingServiceListLinearProcessingRecordsProcedure,
		agentsv1connect.LinearProcessingServiceGetLinearProcessingRecordProcedure,
		agentsv1connect.LinearProcessingServiceResendLinearReplyProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", procHandler)))
	}

	return &linearFixture{fake: fake, resender: resender}, nil
}

func mustCreateLinearApp(
	ctx context.Context,
	svc *application.LinearAppServiceServer,
	app *agentsv1.LinearApp,
	clientSecret, webhookSecret string,
) *agentsv1.LinearApp {
	req := &agentsv1.CreateLinearAppRequest{App: app}
	if clientSecret != "" {
		req.ClientSecret = proto.String(clientSecret)
	}
	if webhookSecret != "" {
		req.WebhookSecret = proto.String(webhookSecret)
	}
	resp, err := svc.CreateLinearApp(ctx, connect.NewRequest(req))
	if err != nil {
		panic(err)
	}
	return resp.Msg.GetApp()
}

// completeLinearInstallCallback hits the public OAuth route for contract tests.
func completeLinearInstallCallback(baseURL, authorizeURL, code string) error {
	u, err := url.Parse(authorizeURL)
	if err != nil {
		return err
	}
	state := u.Query().Get("state")
	if state == "" {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/linear/oauth/callback?state="+url.QueryEscape(state)+"&code="+url.QueryEscape(code), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		return err
	}
	return nil
}

func init() {
	registerFixtureWorkspaces(
		fixtureWorkspace{ID: linearWorkspace, Owner: true, NoSeedAgent: true},
		fixtureWorkspace{ID: linearMemberWorkspace, NoSeedAgent: true},
	)
}
