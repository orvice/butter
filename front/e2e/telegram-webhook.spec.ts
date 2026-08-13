import { expect, test, type Page, type Route } from '@playwright/test'
import { fromBinary } from '@bufbuild/protobuf'
import {
  GetTelegramChannelResponseSchema,
  GetTelegramChannelStatusResponseSchema,
  GetTelegramSettingsResponseSchema,
  ListTelegramDestinationsResponseSchema,
  TelegramCredentialState,
  TelegramReceiveMode,
  TelegramWebhookState,
  UpdateTelegramSettingsRequestSchema,
  UpdateTelegramSettingsResponseSchema,
} from '../src/gen/agents/v1/telegram_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const CHANNEL = {
  id: 'ch-1',
  key: 'ops-bot',
  name: 'Ops bot',
  botId: '111111',
  botUsername: 'opsbot',
  receiveMode: TelegramReceiveMode.WEBHOOK,
  credentialState: TelegramCredentialState.VALID,
  inboundEnabled: true,
  outboundEnabled: true,
  webhookSecretSet: true,
  revision: 1n,
}

async function setupWebhookRoutes(
  page: Page,
  status: Record<string, unknown>,
  overrides: { onUpdateSettings?: (route: Route) => Promise<boolean>; baseUrl?: string } = {}
) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramAdminService/UpdateTelegramSettings')) {
      if (overrides.onUpdateSettings) return overrides.onUpdateSettings(route)
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(UpdateTelegramSettingsRequestSchema, body)
      return fulfillProto(route, UpdateTelegramSettingsResponseSchema, {
        settings: req.settings,
      })
    }
    if (url.includes('TelegramAdminService/GetTelegramSettings')) {
      return fulfillProto(route, GetTelegramSettingsResponseSchema, {
        settings: { webhookBaseUrl: overrides.baseUrl ?? '' },
      })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannelStatus')) {
      return fulfillProto(route, GetTelegramChannelStatusResponseSchema, { status })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannel')) {
      return fulfillProto(route, GetTelegramChannelResponseSchema, { channel: CHANNEL })
    }
    if (url.includes('TelegramDestinationService/ListTelegramDestinations')) {
      return fulfillProto(route, ListTelegramDestinationsResponseSchema, { destinations: [] })
    }
    return false
  })
}

test('an admin configures the public webhook base URL', async ({ page }) => {
  let sent = ''
  await setupWebhookRoutes(page, { channelId: 'ch-1' }, {
    onUpdateSettings: async (route) => {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      sent = fromBinary(UpdateTelegramSettingsRequestSchema, body).settings?.webhookBaseUrl ?? ''
      return fulfillProto(route, UpdateTelegramSettingsResponseSchema, {
        settings: { webhookBaseUrl: sent },
      })
    },
  })

  await page.goto('/admin/telegram')
  await page.getByLabel('Public base URL').fill('https://butter.example.com')
  await page.getByRole('button', { name: 'Save' }).click()

  await expect(page.getByText('Telegram settings updated').first()).toBeVisible()
  expect(sent).toBe('https://butter.example.com')
})

// Telegram only delivers over TLS, so a plain-HTTP base URL would produce a
// registration that silently never works.
test('rejects a non-https base URL', async ({ page }) => {
  await setupWebhookRoutes(page, { channelId: 'ch-1' }, {
    onUpdateSettings: (route) =>
      fulfillConnectError(
        route,
        'invalid_argument',
        'settings.webhook_base_url must use https: Telegram only delivers webhooks over TLS'
      ),
  })

  await page.goto('/admin/telegram')
  await page.getByLabel('Public base URL').fill('http://butter.example.com')
  await page.getByRole('button', { name: 'Save' }).click()

  await expect(page.getByText(/must use https/).first()).toBeVisible()
})

test('shows the derived callback URL and registration state', async ({ page }) => {
  await setupWebhookRoutes(page, {
    channelId: 'ch-1',
    inboundDesired: true,
    outboundDesired: true,
    receiveMode: TelegramReceiveMode.WEBHOOK,
    webhookState: TelegramWebhookState.REGISTERED,
    webhookUrl: 'https://butter.example.com/api/telegram/webhook/ch-1',
    queueReady: true,
  })

  await page.goto('/telegram-channels/ch-1')

  await expect(page.getByTestId('webhook-state')).toContainText('Webhook registered')
  await expect(
    page.getByText('https://butter.example.com/api/telegram/webhook/ch-1')
  ).toBeVisible()
})

test('reports why a webhook channel cannot be enabled', async ({ page }) => {
  await setupWebhookRoutes(page, {
    channelId: 'ch-1',
    receiveMode: TelegramReceiveMode.WEBHOOK,
    webhookState: TelegramWebhookState.PENDING,
    blockers: [
      'no public webhook base URL is configured; a global administrator must set one',
      'redis is not configured as a durable update queue, which webhook mode requires',
    ],
  })

  await page.goto('/telegram-channels/ch-1')

  const blockers = page.getByTestId('channel-blockers')
  await expect(blockers).toContainText('no public webhook base URL is configured')
  await expect(blockers).toContainText('durable update queue')
})
