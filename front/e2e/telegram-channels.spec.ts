import { expect, test, type Page, type Route } from '@playwright/test'
import { fromBinary, type MessageInitShape } from '@bufbuild/protobuf'
import {
  CreateTelegramChannelRequestSchema,
  CreateTelegramChannelResponseSchema,
  GetTelegramChannelResponseSchema,
  GetTelegramChannelStatusResponseSchema,
  ListTelegramChannelsResponseSchema,
  ListTelegramDestinationsResponseSchema,
  RotateTelegramChannelCredentialRequestSchema,
  RotateTelegramChannelCredentialResponseSchema,
  TelegramCredentialState,
  TelegramReceiveMode,
  type TelegramChannelSchema,
} from '../src/gen/agents/v1/telegram_pb'
import {
  fulfillConnectError,
  fulfillProto,
  setupAuthenticatedConnectRoutes,
} from './support/connect'

const CHANNEL: MessageInitShape<typeof TelegramChannelSchema> = {
  id: 'ch-1',
  key: 'ops-bot',
  name: 'Ops bot',
  botId: '111111',
  botUsername: 'opsbot',
  receiveMode: TelegramReceiveMode.WEBHOOK,
  credentialState: TelegramCredentialState.VALID,
  revision: 1n,
}

function decodeRequest<T extends Parameters<typeof fromBinary>[0]>(
  schema: T,
  route: Route
) {
  const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
  return fromBinary(schema, body)
}

async function setupTelegramRoutes(
  page: Page,
  overrides: {
    onCreate?: (route: Route) => Promise<boolean>
    onRotate?: (route: Route) => Promise<boolean>
    blockers?: string[]
    destinations?: unknown[]
  } = {}
) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramChannelService/ListTelegramChannels')) {
      return fulfillProto(route, ListTelegramChannelsResponseSchema, {
        channels: [CHANNEL],
      })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannelStatus')) {
      return fulfillProto(route, GetTelegramChannelStatusResponseSchema, {
        status: {
          channelId: 'ch-1',
          credentialState: TelegramCredentialState.VALID,
          receiveMode: TelegramReceiveMode.WEBHOOK,
          blockers: overrides.blockers ?? [],
          inboundDestinationCount: 0,
          outboundDestinationCount: 0,
        },
      })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannel')) {
      return fulfillProto(route, GetTelegramChannelResponseSchema, {
        channel: CHANNEL,
      })
    }
    if (url.includes('TelegramChannelService/CreateTelegramChannel')) {
      if (overrides.onCreate) return overrides.onCreate(route)
      return fulfillProto(route, CreateTelegramChannelResponseSchema, {
        channel: CHANNEL,
      })
    }
    if (url.includes('TelegramChannelService/RotateTelegramChannelCredential')) {
      if (overrides.onRotate) return overrides.onRotate(route)
      return fulfillProto(route, RotateTelegramChannelCredentialResponseSchema, {
        channel: CHANNEL,
      })
    }
    if (url.includes('TelegramDestinationService/ListTelegramDestinations')) {
      return fulfillProto(route, ListTelegramDestinationsResponseSchema, {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        destinations: (overrides.destinations ?? []) as any,
      })
    }
    return false
  })
}

test('lists registered bots with their pinned identity', async ({ page }) => {
  await setupTelegramRoutes(page)

  await page.goto('/telegram-channels')

  const card = page.getByTestId('telegram-channel-ops-bot')
  await expect(card).toContainText('Ops bot')
  await expect(card).toContainText('@opsbot')
  await expect(card).toContainText('111111')
  await expect(card).toContainText('Webhook')
  // A freshly created channel is disabled until enablement passes preflight.
  await expect(card).toContainText('Disabled')
})

test('sends the bot token write-only when creating a channel', async ({ page }) => {
  let sentToken = ''
  await setupTelegramRoutes(page, {
    onCreate: async (route) => {
      const req = decodeRequest(CreateTelegramChannelRequestSchema, route)
      sentToken = req.botToken
      return fulfillProto(route, CreateTelegramChannelResponseSchema, {
        channel: CHANNEL,
      })
    },
  })

  await page.goto('/telegram-channels/create')
  await page.getByLabel('Key').fill('ops-bot')
  await page.getByLabel('Bot token').fill('111111:secret-token')
  await page.getByRole('button', { name: 'Validate and save' }).click()

  await expect(page).toHaveURL(/\/telegram-channels\/ch-1/)
  expect(sentToken).toBe('111111:secret-token')
  // The token never comes back: the detail page has no field showing it.
  await expect(page.getByText('111111:secret-token')).toHaveCount(0)
})

test('surfaces a rejected bot token as a validation error', async ({ page }) => {
  await setupTelegramRoutes(page, {
    onCreate: (route) =>
      fulfillConnectError(route, 'invalid_argument', 'bot_token was rejected by Telegram'),
  })

  await page.goto('/telegram-channels/create')
  await page.getByLabel('Key').fill('ops-bot')
  await page.getByLabel('Bot token').fill('bad-token')
  await page.getByRole('button', { name: 'Validate and save' }).click()

  await expect(page.getByText(/rejected by Telegram/).first()).toBeVisible()
  await expect(page).toHaveURL(/\/telegram-channels\/create/)
})

test('shows why a channel cannot be enabled yet', async ({ page }) => {
  await setupTelegramRoutes(page, {
    blockers: ['no inbound destination is enabled for this channel'],
  })

  await page.goto('/telegram-channels/ch-1')

  await expect(page.getByTestId('channel-blockers')).toContainText(
    'no inbound destination is enabled'
  )
})

test('rotating to a different bot is reported without losing the current token', async ({
  page,
}) => {
  let rotatedWith = ''
  await setupTelegramRoutes(page, {
    onRotate: async (route) => {
      rotatedWith = decodeRequest(
        RotateTelegramChannelCredentialRequestSchema,
        route
      ).botToken
      return fulfillConnectError(
        route,
        'invalid_argument',
        'bot_token resolves to bot 222222 but this channel is pinned to bot 111111'
      )
    },
  })

  await page.goto('/telegram-channels/ch-1')
  await page.getByLabel('New bot token').fill('222222:other-bot')
  await page.getByRole('button', { name: 'Rotate' }).click()

  await expect(page.getByText(/pinned to bot 111111/).first()).toBeVisible()
  expect(rotatedWith).toBe('222222:other-bot')
  // The channel keeps reporting a valid credential.
  await expect(page.getByText('Token valid').first()).toBeVisible()
})

test('marks a forum topic destination as a distinct address', async ({ page }) => {
  await setupTelegramRoutes(page, {
    destinations: [
      {
        id: 'd-1',
        key: 'general',
        name: 'General',
        channelId: 'ch-1',
        chatId: '-1001234567890',
        messageThreadId: '',
        outboundEnabled: true,
        revision: 1n,
      },
      {
        id: 'd-2',
        key: 'incidents',
        name: 'Incidents',
        channelId: 'ch-1',
        chatId: '-1001234567890',
        messageThreadId: '42',
        inboundEnabled: true,
        outboundEnabled: true,
        revision: 1n,
      },
    ],
  })

  await page.goto('/telegram-channels/ch-1')

  await expect(page.getByTestId('telegram-destination-general')).toContainText('no topic')
  await expect(page.getByTestId('telegram-destination-incidents')).toContainText(
    'topic 42'
  )
})
