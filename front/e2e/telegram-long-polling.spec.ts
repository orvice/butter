import { expect, test, type Page } from '@playwright/test'
import {
  GetTelegramChannelResponseSchema,
  GetTelegramChannelStatusResponseSchema,
  ListTelegramDestinationsResponseSchema,
  TelegramCredentialState,
  TelegramReceiveMode,
} from '../src/gen/agents/v1/telegram_pb'
import { fulfillProto, setupAuthenticatedConnectRoutes } from './support/connect'

function channel(receiveMode: TelegramReceiveMode) {
  return {
    id: 'ch-1',
    key: 'ops-bot',
    name: 'Ops bot',
    botId: '111111',
    botUsername: 'opsbot',
    receiveMode,
    credentialState: TelegramCredentialState.VALID,
    inboundEnabled: true,
    outboundEnabled: true,
    revision: 1n,
  }
}

async function setupPollingRoutes(
  page: Page,
  receiveMode: TelegramReceiveMode,
  status: Record<string, unknown>
) {
  await setupAuthenticatedConnectRoutes(page, async (route, url) => {
    if (url.includes('TelegramChannelService/GetTelegramChannelStatus')) {
      return fulfillProto(route, GetTelegramChannelStatusResponseSchema, { status })
    }
    if (url.includes('TelegramChannelService/GetTelegramChannel')) {
      return fulfillProto(route, GetTelegramChannelResponseSchema, {
        channel: channel(receiveMode),
      })
    }
    if (url.includes('TelegramDestinationService/ListTelegramDestinations')) {
      return fulfillProto(route, ListTelegramDestinationsResponseSchema, { destinations: [] })
    }
    return false
  })
}

test('shows long polling leadership and progress', async ({ page }) => {
  await setupPollingRoutes(page, TelegramReceiveMode.LONG_POLLING, {
    channelId: 'ch-1',
    receiveMode: TelegramReceiveMode.LONG_POLLING,
    pollingLeader: true,
    lastFetchedUpdateId: 205n,
    lastAcceptedUpdateId: 204n,
    queueReady: true,
  })

  await page.goto('/telegram-channels/ch-1')

  await expect(page.getByTestId('polling-state')).toContainText('Polling on this pod')
  // Fetched and accepted are shown separately: the gap is what an operator
  // needs when updates arrive but nothing happens.
  await expect(page.getByTestId('polling-progress')).toContainText('fetched 205')
  await expect(page.getByTestId('polling-progress')).toContainText('accepted 204')
  await expect(page.getByTestId('webhook-state')).toHaveCount(0)
})

test('reports when another pod holds the polling lease', async ({ page }) => {
  await setupPollingRoutes(page, TelegramReceiveMode.LONG_POLLING, {
    channelId: 'ch-1',
    receiveMode: TelegramReceiveMode.LONG_POLLING,
    pollingLeader: false,
  })

  await page.goto('/telegram-channels/ch-1')

  await expect(page.getByTestId('polling-state')).toContainText('another pod')
})

test('reports long polling prerequisites as blockers', async ({ page }) => {
  await setupPollingRoutes(page, TelegramReceiveMode.LONG_POLLING, {
    channelId: 'ch-1',
    receiveMode: TelegramReceiveMode.LONG_POLLING,
    blockers: [
      'redis is not configured, which long polling requires for the update queue and the consumer lease',
    ],
  })

  await page.goto('/telegram-channels/ch-1')

  await expect(page.getByTestId('channel-blockers')).toContainText(
    'long polling requires'
  )
})

// Switching receive mode has real semantics; the form says so before saving.
test('warns before switching receive mode', async ({ page }) => {
  await setupPollingRoutes(page, TelegramReceiveMode.WEBHOOK, { channelId: 'ch-1' })

  await page.goto('/telegram-channels/ch-1')
  await expect(page.getByTestId('mode-switch-warning')).toHaveCount(0)

  await page.getByRole('combobox', { name: 'Receive mode' }).click()
  await page.getByRole('option', { name: 'Long polling' }).click()

  await expect(page.getByTestId('mode-switch-warning')).toContainText(
    'does not guarantee a lossless or duplicate-free transition'
  )
})
