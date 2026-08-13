import { expect, test, type Route } from '@playwright/test'
import { fromBinary } from '@bufbuild/protobuf'
import {
  CreateTelegramChannelRequestSchema,
  CreateTelegramChannelResponseSchema,
  CreateTelegramDestinationRequestSchema,
  CreateTelegramDestinationResponseSchema,
  GetTelegramChannelResponseSchema,
  GetTelegramChannelStatusResponseSchema,
  GetTelegramSettingsResponseSchema,
  ListTelegramChannelsResponseSchema,
  ListTelegramDestinationsResponseSchema,
  ListTelegramProcessingRecordsResponseSchema,
  ResendTelegramReplyResponseSchema,
  SendTelegramTestMessageResponseSchema,
  SetTelegramChannelEnabledRequestSchema,
  SetTelegramChannelEnabledResponseSchema,
  TelegramCredentialState,
  TelegramProcessingStatus,
  TelegramReceiveMode,
  TelegramWebhookState,
  UpdateTelegramSettingsRequestSchema,
  UpdateTelegramSettingsResponseSchema,
} from '../src/gen/agents/v1/telegram_pb'
import { ListAgentsResponseSchema } from '../src/gen/agents/v1/agent_service_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

// The complete owner journey the cutover has to support end to end (#273):
// register a bot, point it at a forum topic, prove the address works, turn
// production receive on, diagnose status, and recover a failed reply.
test('an owner goes from bot creation to failure recovery', async ({ page }) => {
  const state = {
    baseUrl: '',
    createdChannel: false,
    createdDestination: false,
    enabled: false,
    tested: false,
    resent: false,
  }

  const channel = () => ({
    id: 'ch-1',
    key: 'ops-bot',
    name: 'Ops bot',
    botId: '111111',
    botUsername: 'opsbot',
    receiveMode: TelegramReceiveMode.WEBHOOK,
    credentialState: TelegramCredentialState.VALID,
    webhookSecretSet: true,
    inboundEnabled: state.enabled,
    outboundEnabled: state.enabled,
    revision: 1n,
  })

  const destination = () => ({
    id: 'dest-1',
    key: 'incidents',
    name: 'Incidents',
    channelId: 'ch-1',
    chatId: '-1001234567890',
    messageThreadId: '42',
    inboundEnabled: true,
    outboundEnabled: true,
    revision: 1n,
    config: { agentId: 'support' },
    verification: { verified: state.tested },
  })

  await setupAuthenticatedConnectRoutes(page, async (route: Route, url: string) => {
    if (url.includes('TelegramAdminService/UpdateTelegramSettings')) {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      state.baseUrl = fromBinary(UpdateTelegramSettingsRequestSchema, body).settings?.webhookBaseUrl ?? ''
      return fulfillProto(route, UpdateTelegramSettingsResponseSchema, {
        settings: { webhookBaseUrl: state.baseUrl },
      })
    }
    if (url.includes('TelegramAdminService/GetTelegramSettings')) {
      return fulfillProto(route, GetTelegramSettingsResponseSchema, {
        settings: { webhookBaseUrl: state.baseUrl },
      })
    }
    if (url.includes('TelegramChannelService/CreateTelegramChannel')) {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(CreateTelegramChannelRequestSchema, body)
      // The token is write-only: it goes up, and nothing comes back.
      expect(req.botToken).toBe('111111:secret-token')
      state.createdChannel = true
      return fulfillProto(route, CreateTelegramChannelResponseSchema, { channel: channel() })
    }
    if (url.includes('TelegramChannelService/SetTelegramChannelEnabled')) {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(SetTelegramChannelEnabledRequestSchema, body)
      state.enabled = req.inboundEnabled || req.outboundEnabled
      return fulfillProto(route, SetTelegramChannelEnabledResponseSchema, {
        channel: channel(),
        warnings: ['BotFather Group Privacy is enabled: in groups the bot only receives commands and replies'],
      })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannelStatus')) {
      return fulfillProto(route, GetTelegramChannelStatusResponseSchema, {
        status: {
          channelId: 'ch-1',
          receiveMode: TelegramReceiveMode.WEBHOOK,
          inboundDesired: state.enabled,
          outboundDesired: state.enabled,
          queueReady: true,
          webhookState: state.enabled
            ? TelegramWebhookState.REGISTERED
            : TelegramWebhookState.PENDING,
          webhookUrl: state.baseUrl ? `${state.baseUrl}/api/telegram/webhook/ch-1` : '',
          inboundDestinationCount: state.createdDestination ? 1 : 0,
          blockers: state.createdDestination
            ? []
            : ['no inbound destination is enabled for this channel'],
        },
      })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannel')) {
      return fulfillProto(route, GetTelegramChannelResponseSchema, { channel: channel() })
    }
    if (url.includes('TelegramChannelService/ListTelegramChannels')) {
      return fulfillProto(route, ListTelegramChannelsResponseSchema, {
        channels: state.createdChannel ? [channel()] : [],
      })
    }
    if (url.includes('TelegramDestinationService/CreateTelegramDestination')) {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(CreateTelegramDestinationRequestSchema, body)
      expect(req.destination?.messageThreadId).toBe('42')
      state.createdDestination = true
      return fulfillProto(route, CreateTelegramDestinationResponseSchema, {
        destination: destination(),
      })
    }
    if (url.includes('TelegramDestinationService/ListTelegramDestinations')) {
      return fulfillProto(route, ListTelegramDestinationsResponseSchema, {
        destinations: state.createdDestination ? [destination()] : [],
      })
    }
    if (url.includes('TelegramDestinationService/SendTelegramTestMessage')) {
      state.tested = true
      return fulfillProto(route, SendTelegramTestMessageResponseSchema, {
        destination: destination(),
        messageIds: ['900'],
      })
    }
    if (url.includes('TelegramProcessingService/ResendTelegramReply')) {
      state.resent = true
      return fulfillProto(route, ResendTelegramReplyResponseSchema, {
        record: { id: 'rec-1', status: TelegramProcessingStatus.SUCCEEDED },
        messageIds: ['901'],
      })
    }
    if (url.includes('TelegramProcessingService/ListTelegramProcessingRecords')) {
      return fulfillProto(route, ListTelegramProcessingRecordsResponseSchema, {
        records: [
          {
            id: 'rec-1',
            channelId: 'ch-1',
            destinationId: 'dest-1',
            updateId: 501n,
            status: state.resent
              ? TelegramProcessingStatus.SUCCEEDED
              : TelegramProcessingStatus.FAILED,
            attempts: 2,
            error: 'telegram error 500: Internal Server Error',
            output: 'part one\npart two',
            segments: [
              { index: 0, text: 'part one', status: 'sent', messageId: '900' },
              { index: 1, text: 'part two', status: state.resent ? 'sent' : 'failed' },
            ],
          },
        ],
      })
    }
    if (url.includes('AgentService/ListAgents')) {
      return fulfillProto(route, ListAgentsResponseSchema, {
        agents: [{ name: 'Support', agentId: 'support' }],
      })
    }
    return false
  })

  // 1. A global admin sets the public callback host.
  await page.goto('/admin/telegram')
  await page.getByLabel('Public base URL').fill('https://butter.example.com')
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page.getByText('Telegram settings updated').first()).toBeVisible()

  // 2. The owner registers the bot with a write-only token.
  await page.goto('/telegram-channels/create')
  await page.getByLabel('Key').fill('ops-bot')
  await page.getByLabel('Bot token').fill('111111:secret-token')
  await page.getByRole('button', { name: 'Validate and save' }).click()
  await expect(page).toHaveURL(/\/telegram-channels\/ch-1/)

  // 3. Enabling is blocked until an inbound destination exists.
  await expect(page.getByTestId('channel-blockers')).toContainText(
    'no inbound destination is enabled'
  )

  // 4. The owner binds one exact forum topic.
  await page.goto('/telegram-channels/ch-1/destinations/create')
  await page.getByLabel('Key').fill('incidents')
  await page.getByLabel('Chat ID').fill('-1001234567890')
  await page.getByLabel('Topic ID (optional)').fill('42')
  await page.getByRole('combobox', { name: 'Default agent' }).click()
  await page.getByRole('option', { name: 'Support' }).click()
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page).toHaveURL(/\/telegram-destinations\/dest-1/)

  // 5. A test message proves the address before anything is switched on.
  await page.goto('/telegram-channels/ch-1')
  await page.getByLabel('Send test message to incidents').click()
  await expect(page.getByText('Test message delivered').first()).toBeVisible()

  // 6. Turning on production receive surfaces the degraded-capability warning.
  await page.getByLabel('Outbound enabled').click()
  await expect(page.getByText(/Group Privacy/).first()).toBeVisible()

  // 7. Status is diagnosable: the derived callback URL is shown.
  await expect(
    page.getByText('https://butter.example.com/api/telegram/webhook/ch-1')
  ).toBeVisible()

  // 8. A failed reply is recovered by resending, never by re-running.
  await page.goto('/telegram-updates')
  await expect(page.getByTestId('processing-error-rec-1')).toContainText(
    'Internal Server Error'
  )
  await expect(page.getByRole('button', { name: /rerun/i })).toHaveCount(0)
  await page.getByLabel('Resend reply for update 501').click()
  await expect(page.getByText('Reply resent').first()).toBeVisible()
})
