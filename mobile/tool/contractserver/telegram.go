package main

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegramprocessingmemory "go.orx.me/apps/butter/internal/repo/telegramprocessing/memory"
	telegramsettingmemory "go.orx.me/apps/butter/internal/repo/telegramsetting/memory"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi/telegramtest"
	"go.orx.me/apps/butter/internal/telegramsend"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const (
	telegramWorkspace         = "telegram"
	telegramContractBotToken  = "111111:contract-tg-bot"
	telegramContractBotToken2 = "222222:contract-tg-bot-2"
	telegramSeededChannelID   = "tg-seeded-channel"
	telegramSeededDestID      = "tg-seeded-dest"
	telegramResendRecordID    = "tg-resend-record"
)

type contractQueueProbe struct{}

func (contractQueueProbe) Available() bool                       { return true }
func (contractQueueProbe) Ping(context.Context) error            { return nil }
func (contractQueueProbe) CheckReady(context.Context) error      { return nil }
func (contractQueueProbe) CheckLeaseReady(context.Context) error { return nil }

func registerTelegramFixture(
	ctx context.Context,
	r *gin.Engine,
	opts []connect.HandlerOption,
	wsRepo workspacerepo.Repository,
	agentRepo *configmemory.Store,
	repo telegramrepo.Repository,
) error {
	if _, err := agentRepo.CreateAgent(ctx, telegramWorkspace, &agentsv1.Agent{
		AgentId: "tg-agent", Name: "tg-agent", WorkspaceId: telegramWorkspace,
		LifecycleStatus: agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE,
	}); err != nil {
		return err
	}
	if _, err := agentRepo.CreateModelProvider(ctx, telegramWorkspace, &agentsv1.ModelProvider{
		Name: "openai", Type: "openai",
		Models: []*agentsv1.ModelConfig{{Name: "gpt-test", Alias: "fast"}},
	}); err != nil {
		return err
	}

	bots := telegramtest.NewFake()
	bots.Register(telegramContractBotToken, telegramtest.Identity(111111, "contractbot"))
	bots.Register(telegramContractBotToken2, telegramtest.Identity(222222, "contractbot2"))

	keyring := secretbox.NewKeyring(cryptokeymemory.New())

	settings := telegramsettingmemory.New()
	if _, err := settings.Put(ctx, &agentsv1.TelegramSettings{
		WebhookBaseUrl: "https://butter.test",
	}); err != nil {
		return err
	}

	channels := application.NewTelegramChannelServiceServer(repo)
	channels.SetWorkspaceRepo(wsRepo)
	channels.SetKeyring(keyring)
	channels.SetBotFactory(bots.Factory())
	channels.SetSettingsRepo(settings)
	channels.SetQueueProbe(contractQueueProbe{})

	destinations := application.NewTelegramDestinationServiceServer(repo)
	destinations.SetWorkspaceRepo(wsRepo)
	destinations.SetConfigRepos(agentRepo, agentRepo)
	sender := telegramsend.New(repo, keyring, bots.Factory())
	destinations.SetSender(sender)
	destinations.SetReferenceRepos(nil, &memoryCronJobs{jobs: map[string]*agentsv1.CronJob{}})

	processingRepo := telegramprocessingmemory.New()
	processing := application.NewTelegramProcessingServiceServer(processingRepo)
	processing.SetWorkspaceRepo(wsRepo)
	processing.SetSender(sender)

	admin := application.NewTelegramAdminServiceServer(settings)

	seedToken, seedKeyID, err := keyring.Encrypt(ctx, []byte(telegramContractBotToken))
	if err != nil {
		return err
	}
	if _, err := repo.CreateChannel(ctx, telegramWorkspace,
		&agentsv1.TelegramChannel{
			Id: telegramSeededChannelID, Key: "seeded", Name: "Seeded",
			BotId: "111111", BotUsername: "contractbot",
			ReceiveMode:    agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK,
			InboundEnabled: true, OutboundEnabled: true,
			CredentialState: agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_VALID,
		},
		telegramrepo.ChannelCredentials{
			BotToken: telegramrepo.Credential{Ciphertext: seedToken, KeyID: seedKeyID},
		}); err != nil {
		return err
	}
	if _, err := repo.CreateDestination(ctx, telegramWorkspace, &agentsv1.TelegramDestination{
		Id: telegramSeededDestID, Key: "topic", ChannelId: telegramSeededChannelID,
		ChatId: "-1005550001", MessageThreadId: "77",
		InboundEnabled: true, OutboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{AgentId: "tg-agent", Model: "fast"},
	}); err != nil {
		return err
	}

	now := time.Now().UTC()
	record := &agentsv1.TelegramProcessingRecord{
		Id: telegramResendRecordID, WorkspaceId: telegramWorkspace,
		ChannelId: telegramSeededChannelID, DestinationId: telegramSeededDestID,
		UpdateId: 9001,
		Status:   agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_RECEIVED,
	}
	claimed, _, err := processingRepo.Claim(ctx, record, "seed", now, now.Add(time.Minute))
	if err != nil {
		return err
	}
	claimed.Output = "hello from fixture"
	claimed.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER
	claimed.Segments = []*agentsv1.TelegramDeliverySegment{
		{Index: 0, Text: "hello from fixture", Status: "pending"},
	}
	claimed.CreatedAt = timestamppb.New(now)
	if _, err := processingRepo.UpdateClaimed(ctx, claimed, "seed"); err != nil {
		return err
	}
	if err := processingRepo.ReleaseClaim(ctx, telegramWorkspace, claimed.GetId(), "seed"); err != nil {
		return err
	}

	chPath, chHandler := agentsv1connect.NewTelegramChannelServiceHandler(channels, opts...)
	r.Any("/api"+chPath+"*path", gin.WrapH(http.StripPrefix("/api", chHandler)))
	destPath, destHandler := agentsv1connect.NewTelegramDestinationServiceHandler(destinations, opts...)
	r.Any("/api"+destPath+"*path", gin.WrapH(http.StripPrefix("/api", destHandler)))
	procPath, procHandler := agentsv1connect.NewTelegramProcessingServiceHandler(processing, opts...)
	r.Any("/api"+procPath+"*path", gin.WrapH(http.StripPrefix("/api", procHandler)))
	admPath, admHandler := agentsv1connect.NewTelegramAdminServiceHandler(admin, opts...)
	r.Any("/api"+admPath+"*path", gin.WrapH(http.StripPrefix("/api", admHandler)))
	return nil
}

func init() {
	registerFixtureWorkspaces(
		fixtureWorkspace{ID: telegramWorkspace, Owner: true, NoSeedAgent: true},
	)
}
