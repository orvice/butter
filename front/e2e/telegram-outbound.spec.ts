import { expect, test, type Page, type Route } from '@playwright/test'
import { fromBinary } from '@bufbuild/protobuf'
import {
  GetTelegramChannelResponseSchema,
  GetTelegramChannelStatusResponseSchema,
  ListTelegramDestinationsResponseSchema,
  SendTelegramTestMessageRequestSchema,
  SendTelegramTestMessageResponseSchema,
  TelegramCredentialState,
  TelegramReceiveMode,
} from '../src/gen/agents/v1/telegram_pb'
import {
  CreateNotifyGroupRequestSchema,
  CreateNotifyGroupResponseSchema,
  ListNotifyGroupsResponseSchema,
} from '../src/gen/agents/v1/agent_service_pb'
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
  outboundEnabled: true,
  revision: 1n,
}

const DESTINATIONS = [
  {
    id: 'dest-1',
    key: 'incidents',
    name: 'Incidents',
    channelId: 'ch-1',
    chatId: '-1001234567890',
    messageThreadId: '42',
    outboundEnabled: true,
    revision: 1n,
  },
  {
    id: 'dest-2',
    key: 'inbound-only',
    name: 'Inbound only',
    channelId: 'ch-1',
    chatId: '-1009999',
    inboundEnabled: true,
    revision: 1n,
  },
]

async function setupOutboundRoutes(
  page: Page,
  overrides: { onTest?: (route: Route) => Promise<boolean> } = {}
) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramChannelService/GetTelegramChannelStatus')) {
      return fulfillProto(route, GetTelegramChannelStatusResponseSchema, {
        status: { channelId: 'ch-1', outboundDesired: true },
      })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannel')) {
      return fulfillProto(route, GetTelegramChannelResponseSchema, { channel: CHANNEL })
    }
    if (url.includes('TelegramDestinationService/ListTelegramDestinations')) {
      return fulfillProto(route, ListTelegramDestinationsResponseSchema, {
        destinations: DESTINATIONS,
      })
    }
    if (url.includes('TelegramDestinationService/SendTelegramTestMessage')) {
      if (overrides.onTest) return overrides.onTest(route)
      return fulfillProto(route, SendTelegramTestMessageResponseSchema, {
        destination: { ...DESTINATIONS[0], verification: { verified: true } },
        messageIds: ['9001'],
      })
    }
    if (url.includes('NotifyGroupService/ListNotifyGroups')) {
      return fulfillProto(route, ListNotifyGroupsResponseSchema, { notifyGroups: [] })
    }
    return false
  })
}

test('sends a test message to an outbound destination', async ({ page }) => {
  let requestedDestination = ''
  await setupOutboundRoutes(page, {
    onTest: async (route) => {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      requestedDestination = fromBinary(SendTelegramTestMessageRequestSchema, body).destinationId
      return fulfillProto(route, SendTelegramTestMessageResponseSchema, {
        destination: { ...DESTINATIONS[0], verification: { verified: true } },
        messageIds: ['9001'],
      })
    },
  })

  await page.goto('/telegram-channels/ch-1')
  await page.getByLabel('Send test message to incidents').click()

  await expect(page.getByText('Test message delivered').first()).toBeVisible()
  expect(requestedDestination).toBe('dest-1')
})

// An inbound-only destination cannot receive proactive sends, so the action is
// not offered at all rather than failing after the click.
test('does not offer a test message for an inbound-only destination', async ({ page }) => {
  await setupOutboundRoutes(page)

  await page.goto('/telegram-channels/ch-1')

  await expect(page.getByLabel('Send test message to inbound-only')).toBeDisabled()
})

test('reports an unavailable destination instead of claiming success', async ({ page }) => {
  await setupOutboundRoutes(page, {
    onTest: (route) =>
      fulfillConnectError(
        route,
        'failed_precondition',
        'destination dest-1 has outbound delivery disabled'
      ),
  })

  await page.goto('/telegram-channels/ch-1')
  await page.getByLabel('Send test message to incidents').click()

  await expect(page.getByText(/outbound delivery disabled/).first()).toBeVisible()
})

// The notify group form must offer destinations, never a bot token field.
test('notify groups reference a telegram destination', async ({ page }) => {
  let sentDestination = ''
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramDestinationService/ListTelegramDestinations')) {
      return fulfillProto(route, ListTelegramDestinationsResponseSchema, {
        destinations: DESTINATIONS,
      })
    }
    if (url.includes('NotifyGroupService/CreateNotifyGroup')) {
      const body = route.request().postDataBuffer() ?? Buffer.alloc(0)
      const req = fromBinary(CreateNotifyGroupRequestSchema, body)
      sentDestination = req.notifyGroup?.targets[0]?.telegram?.destinationId ?? ''
      return fulfillProto(route, CreateNotifyGroupResponseSchema, {
        notifyGroup: req.notifyGroup,
      })
    }
    if (url.includes('NotifyGroupService/ListNotifyGroups')) {
      return fulfillProto(route, ListNotifyGroupsResponseSchema, { notifyGroups: [] })
    }
    return false
  })

  await page.goto('/notify-groups/create')
  await page.getByLabel('Name').first().fill('ops')
  await page.getByRole('button', { name: 'Add target' }).click()

  // No raw credential input is offered any more.
  await expect(page.getByLabel('Bot Token')).toHaveCount(0)

  await page.getByRole('combobox', { name: 'Telegram destination' }).click()
  await page.getByRole('option', { name: /Incidents/ }).click()
  await page.getByRole('button', { name: 'Create', exact: true }).click()

  await expect.poll(() => sentDestination).toBe('dest-1')
})
